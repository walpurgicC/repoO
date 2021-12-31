package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/nats-io/stan.go"
	"html/template"
	"net/http"
)

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string  `json:"transaction"`
	RequestId    string  `json:"request_id"`
	Currency     string  `json:"currency"`
	Provider     string  `json:"provider"`
	Amount       float64 `json:"amount"`
	PaymentDt    float64 `json:"payment_dt"`
	Bank         string  `json:"bank"`
	DeliveryCost float64 `json:"delivery_cost"`
	GoodsTotal   float64 `json:"goods_total"`
	CustomFee    float64 `json:"custom_fee"`
}

type Item struct {
	ChrtId      float64 `json:"chrt_id"`
	TrackNumber string  `json:"track_number"`
	Price       float64 `json:"price"`
	Rid         string  `json:"rid"`
	Name        string  `json:"name"`
	Sale        float64 `json:"sale"`
	Size        string  `json:"size"`
	TotalPrice  float64 `json:"total_price"`
	NmId        float64 `json:"nm_id"`
	Brand       string  `json:"brand"`
	Status      float64 `json:"status"`
}

type Order struct {
	OrderUid          string   `json:"order_uid"`
	TrackNumber       string   `json:"track_number"`
	Entry             string   `json:"entry"`
	Delivery          Delivery `json:"delivery"`
	Payment           Payment  `json:"payment"`
	Items             []Item   `json:"items"`
	Locale            string   `json:"locale"`
	InternalSignature string   `json:"internal_signature"`
	CustomerId        string   `json:"customer_id"`
	DeliveryService   string   `json:"delivery_service"`
	ShardKey          string   `json:"shardkey"`
	SmId              float64  `json:"sm_id"`
	DateCreated       string   `json:"date_created"`
	OofShard          string   `json:"oof_shard"`
}

func LoadOrdersFromDb() []Order {
	connStr := "user=admin password=admin dbname=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Connected to Postgress")
	}
	defer db.Close()

	var orders []Order
	var i int
	tx, error := db.Begin()
	if error != nil {
		fmt.Println(error.Error())
	}
	rows, _ := tx.Query("select * from orders")
	for rows.Next() {
		var order Order
		_ = rows.Scan(&order.OrderUid, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature, &order.CustomerId, &order.DeliveryService, &order.ShardKey, &order.SmId, &order.DateCreated, &order.OofShard)
		orders = append(orders, order)
	}
	rows.Close()
	for i = 0; i < len(orders); i++ {
		rows, _ := tx.Query("select delivery.name, delivery.phone, delivery.zip, delivery.city, delivery.address, delivery.region, delivery.email from delivery where customer_id = $1", orders[i].CustomerId)
		for rows.Next() {
			_ = rows.Scan(&orders[i].Delivery.Name, &orders[i].Delivery.Phone, &orders[i].Delivery.Zip, &orders[i].Delivery.City, &orders[i].Delivery.Address, &orders[i].Delivery.Region, &orders[i].Delivery.Email)
		}
		rows.Close()
		rows, _ = tx.Query("select * from payment where transaction = $1", orders[i].OrderUid)
		for rows.Next() {
			_ = rows.Scan(&orders[i].Payment.Transaction, &orders[i].Payment.RequestId, &orders[i].Payment.Currency, &orders[i].Payment.Provider, &orders[i].Payment.Amount, &orders[i].Payment.PaymentDt, &orders[i].Payment.Bank, &orders[i].Payment.DeliveryCost, &orders[i].Payment.GoodsTotal, &orders[i].Payment.CustomFee)
		}
		rows.Close()
		rows, _ = tx.Query("select * from items where track_number = $1", orders[i].TrackNumber)
		for rows.Next() {
			var item Item
			_ = rows.Scan(&item.Rid, &item.ChrtId, &item.TrackNumber, &item.Price, &item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmId, &item.Brand, &item.Status)
			orders[i].Items = append(orders[i].Items, item)
		}
		rows.Close()
	}
	tx.Commit()

	return orders
}

func LoadOrderToDb(order Order) {
	connStr := "user=admin password=admin dbname=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Connected to Postgress")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err.Error())
	}
	_, err = tx.Exec("insert into orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		order.OrderUid, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerId, order.DeliveryService, order.ShardKey, order.SmId, order.DateCreated, order.OofShard)
	if err != nil {
		tx.Rollback()
		fmt.Println(err.Error())
	}
	_, err = tx.Exec("insert into delivery (customer_id, name, phone, zip, city, address, region, email) values ($1, $2, $3, $4, $5, $6, $7, $8)",
		order.CustomerId, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		tx.Rollback()
		fmt.Println(err.Error())
	}
	_, err = tx.Exec("insert into payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		order.Payment.Transaction, order.Payment.RequestId, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		tx.Rollback()
		fmt.Println(err.Error())
	}
	_, err = tx.Exec("insert into items (rid, chrt_id, track_number, price, name, sale, size, total_price, nm_id, brand, status) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		order.Items[0].Rid, order.Items[0].ChrtId, order.Items[0].TrackNumber, order.Items[0].Price, order.Items[0].Name, order.Items[0].Sale, order.Items[0].Size, order.Items[0].TotalPrice, order.Items[0].NmId, order.Items[0].Brand, order.Items[0].Status)
	if err != nil {
		tx.Rollback()
		fmt.Println(err.Error())
	}
	err = tx.Commit()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Transaction Succeeded")
	}
}

func main() {
	orders := LoadOrdersFromDb()
	sc, _ := stan.Connect("test-cluster", "ohYea")
	sub, _ := sc.Subscribe("foo", func(msg *stan.Msg) {
		var order Order
		err := json.Unmarshal(msg.Data, &order)
		if err != nil {
			fmt.Println(err.Error())
		}
		LoadOrderToDb(order)
		orders = append(orders, order)
		var i int
		for i = 0; i < len(orders); i++ {
			fmt.Println(orders[i])
		}
	}, stan.DeliverAllAvailable())
	defer sub.Close()
	defer sc.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "form.html")
	})

	http.HandleFunc("/datatable", func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("id")
		for i := 0; i < len(orders); i++ {
			if orders[i].OrderUid == id {
				tmpl, err := template.ParseFiles("dataTable.html")
				if err != nil {
					fmt.Println(err.Error())
				}
				tmpl.Execute(w, orders[i])
			}
		}
	})

	http.ListenAndServe(":8080", nil)
}
