package mysql

import (
	"database/sql"
	"reflect"
)

const (
	MYSQL_FIELD_TAG = "mysql"
)

func ReflectColNames(ret reflect.Type) []string {
	var colNames []string

	if ret.Kind() == reflect.Ptr {
		ret = ret.Elem()
	}

	for i := 0; i < ret.NumField(); i++ {
		retF := ret.Field(i)

		if retF.Type.Kind() == reflect.Struct {
			colNames = append(colNames, ReflectColNames(retF.Type)...)
		}

		if name, ok := retF.Tag.Lookup(MYSQL_FIELD_TAG); ok {
			colNames = append(colNames, name)
		}
	}

	return colNames
}

func ReflectInsertColValues(rev reflect.Value) []interface{} {
	var colValues []interface{}
	ret := rev.Type()

	for i := 0; i < rev.NumField(); i++ {
		revF := rev.Field(i)
		if revF.Kind() == reflect.Struct {
			colValues = append(colValues, ReflectInsertColValues(revF)...)
		}

		_, ok := ret.Field(i).Tag.Lookup(MYSQL_FIELD_TAG)
		if ok {
			colValues = append(colValues, revF.Interface())
		}
	}

	return colValues
}

func ReflectEntityScanValues(rev reflect.Value) []interface{} {
	var scanValues []interface{}
	ret := rev.Type()

	for i := 0; i < rev.NumField(); i++ {
		revF := rev.Field(i)
		if revF.Kind() == reflect.Struct {
			scanValues = append(scanValues, ReflectEntityScanValues(revF)...)
		}

		_, ok := ret.Field(i).Tag.Lookup(MYSQL_FIELD_TAG)
		if ok {
			scanValues = append(scanValues, revF.Addr().Interface())
		}
	}

	return scanValues
}

func ReflectUpdateItems(refOldV, refNewV reflect.Value, updateFields map[string]bool) []*QueryItem {
	var items []*QueryItem
	refNewT := refNewV.Type()

	for i := 0; i < refNewV.NumField(); i++ {
		refNewVF := refNewV.Field(i)
		if refNewVF.Kind() == reflect.Struct {
			items = append(items, ReflectUpdateItems(refOldV.Field(i), refNewVF, updateFields)...)
		}

		refNewTF := refNewT.Field(i)
		colName, ok := refNewTF.Tag.Lookup(MYSQL_FIELD_TAG)
		if !ok {
			continue
		}
		if v, ok := updateFields[colName]; !ok || !v {
			continue
		}

		nv := refNewVF.Interface()
		if nv != refOldV.Field(i).Interface() {
			items = append(items, NewQueryItem(colName, "", nv))
		}
	}

	return items
}

func ReflectQueryItems(rev reflect.Value, required map[string]bool, conditions map[string]string) []*QueryItem {
	var items []*QueryItem
	ret := rev.Type()

	for i := 0; i < rev.NumField(); i++ {
		revF := rev.Field(i)
		if revF.Kind() == reflect.Struct {
			items = append(items, ReflectQueryItems(revF, required, conditions)...)
		}

		retF := ret.Field(i)
		name, ok := retF.Tag.Lookup(MYSQL_FIELD_TAG)
		if !ok {
			continue
		}
		if v, ok := required[name]; !ok || !v {
			continue
		}
		cond, ok := conditions[name]
		if !ok {
			continue
		}

		items = append(items, NewQueryItem(name, cond, revF.Interface()))
	}

	return items
}

func ReflectQueryRowsToEntityList(rows *sql.Rows, ret reflect.Type, listPtr interface{}) error {
	if rows.Next() == false {
		return nil
	}

	revListV := reflect.ValueOf(listPtr).Elem()
	rev := reflect.New(ret)
	destV := ReflectEntityScanValues(rev.Elem())
	err := rows.Scan(destV...)
	if err != nil {
		return err
	}
	revListV.Set(reflect.Append(revListV, rev))

	for rows.Next() {
		rev = reflect.New(ret)
		destV = ReflectEntityScanValues(rev.Elem())
		err = rows.Scan(destV...)
		if err != nil {
			return err
		}
		revListV.Set(reflect.Append(revListV, rev))
	}

	return nil
}