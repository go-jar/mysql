package mysql

import (
	"database/sql"
	"reflect"
)

const (
	FieldTag = "mysql"
)

func ReflectColNames(ret reflect.Type) []string {
	if ret.Kind() == reflect.Ptr {
		ret = ret.Elem()
	}

	if ret.Kind() != reflect.Struct {
		return nil
	}

	var colNames []string

	for i := 0; i < ret.NumField(); i++ {
		retF := ret.Field(i)

		if retF.Type.Kind() == reflect.Ptr || retF.Type.Kind() == reflect.Struct {
			colNames = append(colNames, ReflectColNames(retF.Type)...)
		}

		if name, ok := retF.Tag.Lookup(FieldTag); ok {
			colNames = append(colNames, name)
		}
	}

	return colNames
}

func ReflectInsertColValues(rev reflect.Value) []interface{} {
	if rev.Type().Kind() == reflect.Ptr {
		rev = rev.Elem()
	}

	if rev.Kind() != reflect.Struct {
		return nil
	}

	var colValues []interface{}

	for i := 0; i < rev.NumField(); i++ {
		revF := rev.Field(i)

		if revF.Kind() == reflect.Ptr || revF.Kind() == reflect.Struct {
			colValues = append(colValues, ReflectInsertColValues(revF)...)
		}

		ret := rev.Type()
		_, ok := ret.Field(i).Tag.Lookup(FieldTag)
		if ok {
			colValues = append(colValues, revF.Interface())
		}
	}

	return colValues
}

func ReflectEntityScanValues(rev reflect.Value) []interface{} {
	if rev.Type().Kind() == reflect.Ptr {
		rev = rev.Elem()
	}

	if rev.Kind() != reflect.Struct {
		return nil
	}

	var scanValues []interface{}
	ret := rev.Type()

	for i := 0; i < rev.NumField(); i++ {
		revF := rev.Field(i)

		if revF.Kind() == reflect.Ptr || revF.Kind() == reflect.Struct {
			scanValues = append(scanValues, ReflectEntityScanValues(revF)...)
		}

		_, ok := ret.Field(i).Tag.Lookup(FieldTag)
		if ok {
			scanValues = append(scanValues, revF.Addr().Interface())
		}
	}

	return scanValues
}

func ReflectUpdateItems(refOldV, refNewV reflect.Value, updateFields map[string]bool) []*QueryItem {
	if refOldV.Type().Kind() == reflect.Ptr {
		refOldV = refOldV.Elem()
	}

	if refNewV.Type().Kind() == reflect.Ptr {
		refNewV = refNewV.Elem()
	}

	if refOldV.Kind() != reflect.Struct || refNewV.Kind() != reflect.Struct {
		return nil
	}

	var items []*QueryItem
	refNewT := refNewV.Type()

	for i := 0; i < refNewV.NumField(); i++ {
		refNewVF := refNewV.Field(i)

		if refNewVF.Kind() == reflect.Ptr || refNewVF.Kind() == reflect.Struct {
			items = append(items, ReflectUpdateItems(refOldV.Field(i), refNewVF, updateFields)...)
		}

		refNewTF := refNewT.Field(i)
		colName, ok := refNewTF.Tag.Lookup(FieldTag)
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
	if rev.Type().Kind() == reflect.Ptr {
		rev = rev.Elem()
	}

	if rev.Kind() != reflect.Struct {
		return nil
	}

	var items []*QueryItem
	ret := rev.Type()

	for i := 0; i < rev.NumField(); i++ {
		revF := rev.Field(i)

		if revF.Kind() == reflect.Ptr || revF.Kind() == reflect.Struct {
			items = append(items, ReflectQueryItems(revF, required, conditions)...)
		}

		retF := ret.Field(i)
		name, ok := retF.Tag.Lookup(FieldTag)
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
	destV := ReflectEntityScanValues(rev)
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
