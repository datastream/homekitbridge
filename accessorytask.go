package main

import (
        "fmt"
        "github.com/datastream/sqlgen"
        "time"
)
type Accessorys struct {
	ID int
	Key string
	Name string
	SerialNumber string
	Manufacturer string
	Model string
	Pin string
	AccessoryType string
}
func accessorysColumns() []string {
	return []string{"id", "key", "name", "serial_number", "manu_facturer","model", "pin", "accessory_type", "created_at", "updated_at"}
}
func (ac *Accessorys)Save() error {
	q := sqlgen.Insert("accessorys")
        q.Columns(accessorysColumns()[1:]...)
        t := time.Now().Format("2006-01-02 15:04:05")
        q.Values([]interface{}{ac.Key, ac.Name, ac.SerialNumber, ac.Manufacturer, ac.Model, ac.Pin, ac.AccessoryType, t, t})
        querystr, args, err := q.ToSQL()
        if err == nil {
                _, err = homekitBridge.db.Exec(sqlgen.PostgresSQLFormat(querystr), args...)
        }
        lo, err := FindAccessoryByKey(ac.Key)
        if err == nil && lo.ID > 0 {
                ac.ID = lo.ID
        }
        return err
}

func FindAccessoryByKey(key string) (*Accessorys, error) {
	var ac Accessorys
	q := sqlgen.Select(accessorysColumns()...)
        q.From("accessorys")
        q.Where("key = ?", key)
        querystr, args, err := q.ToSQL()
        if err == nil {
                err = homekitBridge.db.QueryRow(sqlgen.PostgresSQLFormat(querystr), args...).Scan(&ac.ID, &ac.Key, &ac.Name, &ac.SerialNumber,&ac.Manufacturer,ac.Model, ac.Pin, ac.AccessoryType, &ac.CreatedAt, &ac.UpdatedAt)
        }
	return &ac, err
}

func FindAccessoryByID(ID int) (*Accessorys, error) {
	var ac Accessorys
	q := sqlgen.Select(accessorysColumns()...)
        q.From("accessorys")
        q.Where("id = ?", ID)
        querystr, args, err := q.ToSQL()
        if err == nil {
                err = homekitBridge.db.QueryRow(sqlgen.PostgresSQLFormat(querystr), args...).Scan(&ac.ID, &ac.Key, &ac.Name, &ac.SerialNumber,&ac.Manufacturer,ac.Model, ac.Pin, ac.AccessoryType, &ac.CreatedAt, &ac.UpdatedAt)
        }
	return &ac, error
}

func (ac *Accessorys)Destroy() error {
	q := sqlgen.Delete()
        q.From("accessorys")
        q.Where("id = ?", ac.ID)
        querystr, args, err := q.ToSQL()
        if err == nil {
                _, err = lb.db.Exec(sqlgen.PostgresSQLFormat(querystr), args...)
        }
        return err
}

func FindAllAccessorys() ([]Accessorys, error) {
	var accessorys []Accessorys
	q := sqlgen.Select(accessorysColumns()...)
        q.From("accessorys")
	querystr, args, err := q.ToSQL()
        rst, err := homekitBridge.db.Query(sqlgen.PostgresSQLFormat(querystr), args...)
        if err != nil {
                return accessorys, err
        }
        for rst.Next() {
                var accessory Accessorys
                if err = rst.Scan(&ac.ID, &ac.Key, &ac.Name, &ac.SerialNumber,&ac.Manufacturer,ac.Model, ac.Pin, ac.AccessoryType, &ac.CreatedAt, &ac.UpdatedAt); err != nil {
                        break
                }
                accessorys = append(accessorys, accessory)

        }
        return accessorys, err
}

func (ac *Accessorys)Task() error {
	info := accessory.Info {
		Name: ac.Name,
		SerialNumber: ac.SerialNumber,
		Manufacturer: ac.Manufacturer,
		Model: ac.Model,
	}
	switch ac.AccessoryType {
	case "TemperatureSensor":
		acc := accessory.NewTemperatureSensor(info,5,-100,50,0.1)
		config := hc.Config{Pin: ac.Pin}
		t, err := hc.NewIPTransport(config, acc.Accessory)
		if err != nil {
			log.Panic(err)
		}

		hc.OnTermination(func() {
			t.Stop()
		})
		for {
			acc.TempSensor.CurrentTemperature.SetValue(homekitBridge.cache.Get(ac.Key))
			_, err := FindAccessoryByID(ac.ID)
			if err != nil && err.Error() == "exists" {
				t.Stop()
				return
			}
			time.Sleep(time.Second * 60)
		}
	case "HumiditySensor":
	case "Switch":
	}
}
