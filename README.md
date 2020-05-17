## CRM
CRM written in Golang. (Implemented in minimum mode. Test regime)

ERP evolution planned in future
### Store data
Used SQLite, MongoDB, Redis to store data.
### Metrics and Monitoring
Used by Prometheus to collect the number of JSON/XML API calls.
### Secure
Used HTTPS for login page.
### Search and Find
Used by ElasticSearch to find data.
### REST API
Used JSON and XML API to get and write data.
### Front-end
Uses HTML and some JavaScript.
### Mail
Implemented the function of sending letters to the mail server
### gRPC
Used gRPC to get and write data.
### RabbitMQ
Implemented the function of sending changed and new customers in Queue
### GraphQL
Search by customer ID, example below
```graphql
{
  FindOneRow(Customer_id: "777"){
    Customer_id
    Customer_name
    Customer_type
    Customer_email
  }
}
```
### For Russian ERP system 1C
Для системы 1С реализована обработка для интеграции, используя JSON API (GET, POST запросы)
Так же есть реализация проверки контрагента в базе налоговой через HTTP
