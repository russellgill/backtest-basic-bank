package main

import (
  "fmt"
  "time"
  "net/url"
  "strings"
  "strconv"
  "context"
  "net/http"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
)

var Database string
var Collection string
var Records string
var client *mongo.Client

type Record struct{
	Identifier	  string `bson: Identifier`
	Transaction   string `bson: Transaction`
	Balance       string `bson: Balance`
	Change        string `bson: Change`
	Date          string `bson: Date`
}

type Account struct{
	Identifier    string	`bson: Identifier`
	Balance       string	`bson: Balance`
	Account_type  string	`bson: Type`
}


func generate_record(ident string, balance string, change string, trx string){
	date := time.Now().String()
	record, err := bson.Marshal(Record{ident, balance, trx, change, date})
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database(Database).Collection(Records)
	_ , err = collection.InsertOne(context.TODO(), record)
	if err != nil {
		fmt.Println(err)
	}
}

// handles the url parsing
func url_parser(url_string string) url.Values{
	params, err := url.ParseQuery(url_string)
	if err != nil{
		fmt.Println(err)
	}
	return params
}

// cleans up strings, converts from []string to string
func clean_string(target []string) string{
	return strings.Join(target, "")
}

// manages deposits to the bank
func deposit(res http.ResponseWriter, req *http.Request){
	var result Account

	collection := client.Database(Database).Collection(Collection)
	params := url_parser(req.URL.String())
	filter := bson.D{{"identifier", clean_string(params["account"])}}
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	
	change, err := strconv.ParseFloat(clean_string(params["deposit"]), 64)
	
	if err != nil {
		fmt.Println(err)
	}
	initial, err := strconv.ParseFloat(result.Balance, 64)
	updated := strconv.FormatFloat((initial + change), 'f', -1, 64)
	result.Balance = updated
	if err != nil{
		fmt.Println(err)
	}
	entry, err := bson.Marshal(result)
	_ , err = collection.ReplaceOne(context.TODO(), filter, entry)
	if err != nil {
		fmt.Println(err)
	}
	generate_record(clean_string(params["account"]), updated, "+"+clean_string(params["deposit"]), "deposit")
}

// manages withdrawls from the bank
func withdraw(res http.ResponseWriter, req *http.Request){
	var result Account

	collection := client.Database(Database).Collection(Collection)
	params := url_parser(req.URL.String())
	filter := bson.D{{"identifier", clean_string(params["account"])}}
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	
	change, err := strconv.ParseFloat(clean_string(params["withdrawl"]), 64)
	
	if err != nil {
		fmt.Println(err)
	}

	initial, err := strconv.ParseFloat(result.Balance, 64)
	updated := strconv.FormatFloat((initial - change), 'f', -1, 64)
	result.Balance = updated

	if err != nil{
		fmt.Println(err)
	}
	entry, err := bson.Marshal(result)
	_ , err = collection.ReplaceOne(context.TODO(), filter, entry)
	if err != nil {
		fmt.Println(err)
	}
	generate_record(clean_string(params["account"]), updated, "-"+clean_string(params["withdrawl"]), "withdrawl")
}

// Instantiates a bank with a new entry with manually determined balance
func inst_bank(res http.ResponseWriter, req *http.Request){

	params := url_parser(req.URL.String())

	account := Account{
			clean_string(params["account"]), 
			clean_string(params["balance"]), 
			clean_string(params["type"])}
	
	entry, err := bson.Marshal(account)

	if err != nil{
		fmt.Println(err)
	}

	fmt.Println(account)
	fmt.Println("Done")

	collection := client.Database(Database).Collection(Collection)
	_ , err = collection.InsertOne(context.TODO(), entry)
	if err != nil{

		fmt.Println(err)
	}
	generate_record(clean_string(params["account"]), clean_string(params["balance"]), "+"+clean_string(params["balance"]), "deposit")
}

// allows queries to the account, returns balances
func query(res http.ResponseWriter, req *http.Request){
	var result Account
	params := url_parser(req.URL.String())
	account := params["account"]
	filter := bson.D{{"identifier", clean_string(account)}}
	collection := client.Database(Database).Collection(Collection)
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil{
		fmt.Println(err)
	}
	fmt.Println(result)
}

// instantiates the server
func server(){

	fmt.Println("Establishing server...")
	http.HandleFunc("/deposit", deposit)
	http.HandleFunc("/withdraw", withdraw)
	http.HandleFunc("/query", query)
	http.HandleFunc("/inst", inst_bank)
	fmt.Println("Done")

	err := http.ListenAndServe(":8080", nil)

	if err != nil{
		fmt.Println(err)
	}
}

// instantiates the database
func inst_database(){
	fmt.Println("Instantiate Database")
	ctx, _ := context.WithTimeout(context.Background(),  10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
}

// main loop
func main(){
	Database = "accounts"
	Collection = "testing"
	Records = "records"
	inst_database()
	server()
}
