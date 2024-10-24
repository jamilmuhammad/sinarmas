package test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Response struct {
	Error *Error      `json:"error"`
	Data  interface{} `json:"data"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ModelProduct struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Category string `json:"category"`
}

type Product struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

type ProductDetail struct {
	Id                string           `json:"id"`
	TransactionTotal  int              `json:"transaction_total"`
	TransactionDate   time.Time        `json:"transaction_date"`
	TransactionDetail []ProductDetails `json:"transaction_detail"`
}

type ProductDetails struct {
	Category string    `json:"category"`
	Items    []Product `json:"items"`
}

type PostProduct struct {
	Product []ModelProduct `json:"product"`
}

type interfaceProductService interface {
	PostProduct(ctx context.Context, payload *PostProduct) (*ProductDetail, error)
}

type productService struct {
	r              *mux.Router
	serviceProduct interfaceProductService
}

func NewService() *productService {
	return &productService{}
}

func RandomString(n int) string {
	randomString := fmt.Sprintf("TR%d", rand.IntN(n))
	return randomString
}

func (s *productService) PostProduct(ctx context.Context, payload *PostProduct) (*ProductDetail, error) {

	log.Println(payload)
	categoryMap := make(map[string][]Product)
	var total = 0
	for _, product := range payload.Product {
		item := Product{
			ID:    product.Id,
			Name:  product.Name,
			Price: product.Price,
		}

		total += product.Price

		categoryMap[product.Category] = append(categoryMap[product.Category], item)
	}

	var transactionDetail []ProductDetails
	for category, items := range categoryMap {
		sort.Slice(items, func(i, j int) bool {
			return items[i].ID < items[j].ID
		})

		transactionDetail = append(transactionDetail, ProductDetails{
			Category: category,
			Items:    items,
		})
	}

	responseProduct := &ProductDetail{
		Id:                RandomString(3),
		TransactionTotal:  total,
		TransactionDate:   time.Now(),
		TransactionDetail: transactionDetail,
	}

	log.Println(responseProduct)

	return responseProduct, nil
}

type productController struct {
	serviceProduct interfaceProductService
	r              *mux.Router
	validate       *validator.Validate
}

func NewController(serviceProduct interfaceProductService, r *mux.Router) *productController {
	return &productController{serviceProduct: serviceProduct, r: r, validate: validator.New()}
}

func (c *productController) PostProduct(w http.ResponseWriter, r *http.Request) {
	var payload PostProduct

	log.Println(r.Body, "r.body")

	//payloads := &PostProduct{
	//	[]Product[
	//		{
	//			"id": 1,
	//			"name": "Computer",
	//			"price": 10000,
	//			"category": "Electronic"
	//		},
	//		{
	//			"id": 2,
	//			"name": "Soap",
	//			"price": 20000,
	//			"category": "Body Care"
	//		},
	//		{
	//			"id": 3,
	//			"name": "Handphone",
	//			"price": 30000,
	//			"category": "Electronic"
	//		},
	//		{
	//			"id": 4,
	//			"name": "Sprite",
	//			"price": 40000,
	//			"category": "Beverage"
	//		},
	//		{
	//			"id": 5,
	//			"name": "Fanta",
	//			"price": 50000,
	//			"category": "Beverage"
	//		},
	//		{
	//			"id": 6,
	//			"name": "Laptop",
	//			"price": 80000,
	//			"category": "Electronic"
	//		}
	//	]
	//}

	err := json.NewDecoder(r.Body).Decode(&payload)

	log.Println(payload, "payload")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		var resp Response
		w.WriteHeader(http.StatusBadRequest)
		resp.Error = &Error{Code: http.StatusBadRequest, Message: "error validate"}
		resp.Data = nil
		res, _ := json.Marshal(resp)
		w.Write(res)
	}

	ctx := r.Context()

	err = c.validate.StructCtx(ctx, payload)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		var resp Response
		resp.Error = &Error{Code: http.StatusBadRequest, Message: "error validate"}
		resp.Data = nil
		res, _ := json.Marshal(resp)
		w.Write(res)
	}

	result, err := c.serviceProduct.PostProduct(ctx, &payload)

	log.Println(result, "result")

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		var resp Response
		resp.Error = &Error{Code: http.StatusBadRequest, Message: "error validate"}
		resp.Data = nil
		res, _ := json.Marshal(resp)
		w.Write(res)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	var resp Response
	resp.Data = result
	resp.Error = nil
	log.Println(resp, "ini apaaa")
	//res, _ := json.Marshal(resp)
	res, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		var resp Response
		resp.Error = &Error{Code: http.StatusInternalServerError, Message: "error validate"}
		resp.Data = nil
		return
	}
	w.Write(res)
}

type server struct {
	mux *mux.Router
}

func NewServer(mux *mux.Router) *server {
	return &server{mux}
}

func (d *productController) Routes() {

	d.r.HandleFunc("/api/v1/product", d.PostProduct).Methods("POST")
	//d.r.HandleFunc("/api/v1/users/auth", d.Login).Methods("POST")
	//d.r.HandleFunc("/api/v1/users/auth/logout", AuthAdminMiddleware(d.Logout)).Methods("POST")
	//d.r.HandleFunc("/api/v1/users/auth/refresh-token", AuthAdminMiddleware(d.RefreshToken)).Methods("POST")
	//d.r.HandleFunc("/api/v1/users", d.GetAllUsers).Methods("GET")
	//d.r.HandleFunc("/api/v1/users/{id}", AuthAdminMiddleware(d.GetUserById)).Methods("GET")
	//d.r.HandleFunc("/api/v1/users/{id}", AuthAdminMiddleware(d.UpdateUser)).Methods("PUT")
	//d.r.HandleFunc("/api/v1/users/{id}", AuthAdminMiddleware(d.DeleteUser)).Methods("DELETE")
	//d.r.HandleFunc("/api/v1/users/{id}/verified", AuthAdminMiddleware(d.VerifiedUser)).Methods("PUT")
	//d.r.HandleFunc("/api/v1/users/{id}/rejected", AuthAdminMiddleware(d.RejectedUser)).Methods("PUT")
}

func (s *server) server() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	useService := NewService()
	useController := NewController(useService, s.mux)
	useController.Routes()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	port := ":9010"

	go func() {
		server := &http.Server{
			Addr:         port,
			Handler:      s.mux,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		}

		err := server.ListenAndServe()
		if err != nil {
			log.Println(err)
			cancel()
		}
	}()

	log.Println("API Gateway listen on port", port)

	select {
	case q := <-quit:
		log.Println("signal.Notify:", q)
	case done := <-ctx.Done():
		log.Println("ctx.Done:", done)
	}

	log.Println("Server API Gateway Exited Properly")

	return nil
}

func main() {
	r := mux.NewRouter()

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:9010"},
		AllowedMethods: []string{"POST"},
		Debug:          true,
	})

	r.Use(corsMiddleware.Handler)

	s := NewServer(r)

	log.Fatal(s.server())

}
