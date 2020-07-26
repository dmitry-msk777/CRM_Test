package gormpostgresql

import (
	RootSctuct "github.com/dmitry-msk777/CRM_Test/RootDescription"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Customer_struct struct {
	gorm.Model
	Customer_id    string
	Customer_name  string
	Customer_type  string
	Customer_email string
}

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

var connected bool
var db *gorm.DB

func CreateConnect(Global_settings RootSctuct.Global_settings) error {
	dbConn, err := gorm.Open(Global_settings.GORM_DataType, Global_settings.GORM_ConnectString)
	if err != nil {
		return err
	}

	db = dbConn
	connected = true
	return nil
}

func GetAllCustomer(Global_settings RootSctuct.Global_settings) ([]RootSctuct.Customer_struct, error) {

	var Customer_struct_slice []Customer_struct

	if connected == false {
		err := CreateConnect(Global_settings)
		if err != nil {
			return nil, err
		}
	}

	db.Find(&Customer_struct_slice)

	Customer_struct_slice_return := []RootSctuct.Customer_struct{}

	for _, customer_gorm := range Customer_struct_slice {

		p := RootSctuct.Customer_struct{Customer_id: customer_gorm.Customer_id,
			Customer_name:  customer_gorm.Customer_name,
			Customer_type:  customer_gorm.Customer_type,
			Customer_email: customer_gorm.Customer_email}

		Customer_struct_slice_return = append(Customer_struct_slice_return, p)
	}

	return Customer_struct_slice_return, nil

}

func AddChangeOneRow(Customer_struct_ext RootSctuct.Customer_struct, Global_settings RootSctuct.Global_settings) error {

	if connected == false {
		err := CreateConnect(Global_settings)
		if err != nil {
			return err
		}
	}

	var Customer_struct_find Customer_struct
	db.First(&Customer_struct_find, "Customer_id = ?", Customer_struct_ext.Customer_id)

	if Customer_struct_find.Customer_id == "" {

		Customer_struct_create := Customer_struct{
			Customer_id:    Customer_struct_ext.Customer_id,
			Customer_name:  Customer_struct_ext.Customer_name,
			Customer_type:  Customer_struct_ext.Customer_type,
			Customer_email: Customer_struct_ext.Customer_email}

		db.Create(&Customer_struct_create)

	} else {
		Customer_struct_upadate := &Customer_struct{}
		Customer_struct_upadate.Customer_id = Customer_struct_ext.Customer_id
		Customer_struct_upadate.Customer_name = Customer_struct_ext.Customer_name
		Customer_struct_upadate.Customer_type = Customer_struct_ext.Customer_type
		Customer_struct_upadate.Customer_email = Customer_struct_ext.Customer_email

		db.Model(&Customer_struct_find).Update(Customer_struct_ext)
	}

	return nil

}

func DeleteOneRow(id string, Global_settings RootSctuct.Global_settings) error {

	if connected == false {
		err := CreateConnect(Global_settings)
		if err != nil {
			return err
		}
	}

	db.Where("Customer_id = ?", id).Delete(&Customer_struct{})

	return nil

}

func FindOneRow(id string, Global_settings RootSctuct.Global_settings) (RootSctuct.Customer_struct, error) {

	if connected == false {
		_ = CreateConnect(Global_settings)
	}

	var Customer_struct Customer_struct
	db.First(&Customer_struct, "Customer_id = ?", id)

	Customer_struct_return := RootSctuct.Customer_struct{}
	Customer_struct_return.Customer_id = Customer_struct.Customer_id
	Customer_struct_return.Customer_name = Customer_struct.Customer_name
	Customer_struct_return.Customer_type = Customer_struct.Customer_type
	Customer_struct_return.Customer_email = Customer_struct.Customer_email

	return Customer_struct_return, nil

}

// // Add Environment Vaable in docker image postgres
// // POSTGRES_PASSWORD password
// // POSTGRES_USER user
// // POSTGRES_DB dbname

//	db,  gorm.Open("postgres", "host=127.0.0.1 port=32768 user=user dbname=dbname password=password sslmode=disable"))
