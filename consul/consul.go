package consul

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	rootsctuct "github.com/dmitry-msk777/CRM_Test/rootdescription"
	consul "github.com/hashicorp/consul/api"
	//enginecrm "github.com/dmitry-msk777/CRM_Test/enginecrm"
)

func TestParsing() {
	// consulAddr := flag.String("consulAddr", "localhost:32769", "a string")
	// flag.Parse()
	// fmt.Println("Consul addres from package consul: ", *consulAddr)
}

func GetSettingsFromConsul(Address string) rootsctuct.Global_settings {

	var consulClient *consul.Client

	consulConf := consul.DefaultConfig()
	consulConf.Address = Address

	var err error
	consulClient, err = consul.NewClient(consulConf)

	if err != nil {
		fmt.Println(err.Error())
	}

	qo := &consul.QueryOptions{
		WaitIndex: 100,
	}

	kvPairs, qm, err := consulClient.KV().List("", qo)

	if err != nil {
		fmt.Println(err.Error())
	}

	// fmt.Println("remoute consul last index", qm.LastIndex)
	// if qm.LastIndex == 1000 {
	// 	fmt.Println("Consult not changed")
	// }

	fmt.Println("qm", qm)

	//newConfig := make(map[string]string)

	GlobalSettingsReturn := rootsctuct.Global_settings{}

	PrifixGroup := "GlobalSetting/"

	reflectGlobalSettings := reflect.ValueOf(&GlobalSettingsReturn)
	reflectElem := reflectGlobalSettings.Elem()

	for _, item := range kvPairs {
		if item.Key == PrifixGroup {
			continue
		}
		//fmt.Println(string(item.Key), string(item.Value))
		res := strings.ReplaceAll(string(item.Key), PrifixGroup, "")
		//fmt.Println("res:", res)

		field := reflectElem.FieldByName(res)

		if field.IsValid() {

			switch field.Kind() {
			case reflect.Int:
				{
					ParseIntVariable, _ := strconv.Atoi(string(item.Value))
					field.SetInt(int64(ParseIntVariable))
				}
			case reflect.Bool:
				{
					ParseBoolVariable, _ := strconv.ParseBool(string(item.Value))
					field.SetBool(ParseBoolVariable)
				}
			case reflect.String:
				{
					field.SetString(string(item.Value))
				}
			}

		}
	}

	//fmt.Println("value res:", GlobalSettingsReturn)

	return GlobalSettingsReturn

}
