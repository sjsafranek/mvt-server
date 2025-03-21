package main

import (
	"errors"
	"fmt"
	"strings"
	// "text/template"
)

// import sq "github.com/Masterminds/squirrel"
// // text/template
//
// func init() {
// 	sql, _, _ := sq.SelectBuilder{}.Where("name IN (?,?)", "Dumbo", "Verna").ToSql()
// 	fmt.Println(sql)
// }

type Filters struct {
	Logical    string                 `json:"logical"`
	Conditions []FeatureServiceFilter `json:"conditions"`
}

func (self *Filters) ToSQL(layer *Layer) (string, error) {
	where := ""
	if 0 != len(self.Conditions) {
		where = "WHERE "
		for i := range self.Conditions {
			if 0 != i {
				where += " AND "
			}
			condition, err := self.Conditions[i].ToSQL(layer)
			if nil != err {
				return "", err
			}
			where += condition
		}
	}

	// HACK
	if 0 != strings.Count(where, ";") {
		logger.Critical(where)
		return "", errors.New("Invalid filter")
	}
	//.end

	return where, nil
}

type FeatureServiceQuery struct {
	Method  string                 `json:"method"`
	Limit   int                    `json:"limit,omitempty"`
	Layer   string                 `json:"layer"`
	Filters []FeatureServiceFilter `json:"filters"`
	ToSRID  int64                  `json:"to_srid,omitempty"` // not hooked up
}

func (self *FeatureServiceQuery) ToSQL() (string, error) {

	layer, err := LAYERS.GetLayer(self.Layer)
	if nil != err {
		return "", err
	}

	limit := ""
	if 0 < self.Limit {
		limit = fmt.Sprintf("LIMIT %v", self.Limit)
	}

	where := ""
	if 0 != len(self.Filters) {
		where = "WHERE "
		for i := range self.Filters {
			if 0 != i {
				where += " AND "
			}
			condition, err := self.Filters[i].ToSQL(layer)
			if nil != err {
				return "", err
			}
			where += condition
		}
	}

	if 4269 != layer.SRID {
		return fmt.Sprintf(`
            SELECT
                json_build_object('type', 'FeatureCollection', 'features', json_agg((

                SELECT
                    json_build_object('type', 'Feature', 'geometry', ST_AsGeoJSON( ST_Transform(ST_SetSRID(geom, %v), 4269) )::jsonb, 'properties', row_to_json(lyr)::jsonb - 'geom' )
                FROM
                    "%v" AS lyr
                %v
                %v
            )));
        `, layer.SRID, self.Layer, where, limit), nil
	}

	return fmt.Sprintf(`
        SELECT
            json_build_object('type', 'FeatureCollection', 'features', json_agg((

            SELECT
                json_build_object('type', 'Feature', 'geometry', ST_AsGeoJSON(geom)::jsonb, 'properties', row_to_json(lyr)::jsonb - 'geom' )
            FROM
                "%v" AS lyr
            %v
            %v
        )));
    `, self.Layer, where, limit), nil

}

type FeatureServiceFilter struct {
	Test     string        `json:"test"`
	Wkt      string        `json:"wkt,omitempty"`
	ColumnId string        `json:"column_id,omitempty"`
	Values   []interface{} `json:"values,omitempty"`
	Min      int64         `json:"min,omitempty"`
	Max      int64         `json:"max,omitempty"`
}

func (self *FeatureServiceFilter) ToSQL(layer *Layer) (string, error) {
	// TODO
	//  - sql injection protection...
	// https://www.calhoun.io/what-is-sql-injection-and-how-do-i-avoid-it-in-go/

	switch strings.ToLower(self.Test) {

	case "contains":
		if 4269 != layer.SRID {
			return fmt.Sprintf("ST_Contains(ST_Transform(ST_SetSRID(geom, %v), 4269), ST_SetSRID(ST_GeomFromText('%v'), 4269))", layer.SRID, self.Wkt), nil
		}
		return fmt.Sprintf("ST_Contains(geom, ST_GeomFromText('%v'))", self.Wkt), nil

	case "within":
		if 4269 != layer.SRID {
			return fmt.Sprintf("ST_Within(ST_Transform(ST_SetSRID(geom, %v), 4269), ST_SetSRID(ST_GeomFromText('%v'), 4269))", layer.SRID, self.Wkt), nil
		}
		return fmt.Sprintf("ST_Within(geom, ST_GeomFromText('%v'))", self.Wkt), nil

	case "overlaps":
		if 4269 != layer.SRID {
			return fmt.Sprintf("ST_Overlaps(ST_Transform(ST_SetSRID(geom, %v), 4269), ST_SetSRID(ST_GeomFromText('%v'), 4269))", layer.SRID, self.Wkt), nil
		}
		return fmt.Sprintf("ST_Overlaps(geom, ST_GeomFromText('%v'))", self.Wkt), nil

	case "not":
		column_exists := false

		for i := range layer.Columns {
			if layer.Columns[i].ColumnId == self.ColumnId {
				column_exists = true
				break
			}
		}
		if !column_exists {
			return "", fmt.Errorf("Column not found: %v", self.ColumnId)
		}

		values := []string{}
		for i := range self.Values {
			// json numbers get Unmarshaled to float64??
			if val, ok := self.Values[i].(string); ok {
				values = append(values, fmt.Sprintf("'%v'", val))
			} else {
				// float64
				values = append(values, fmt.Sprintf("%v", self.Values[i]))
			}
		}
		return fmt.Sprintf("NOT ( %v in (%v) )", self.ColumnId, strings.Join(values, ",")), nil

	case "match":
		column_exists := false

		for i := range layer.Columns {
			if layer.Columns[i].ColumnId == self.ColumnId {
				column_exists = true
				break
			}
		}
		if !column_exists {
			return "", fmt.Errorf("Column not found: %v", self.ColumnId)
		}

		values := []string{}
		for i := range self.Values {
			// json numbers get Unmarshaled to float64??
			if val, ok := self.Values[i].(string); ok {
				values = append(values, fmt.Sprintf("'%v'", val))
			} else {
				// float64
				values = append(values, fmt.Sprintf("%v", self.Values[i]))
			}
		}
		return fmt.Sprintf("%v IN (%v)", self.ColumnId, strings.Join(values, ",")), nil

	case "range":
		return fmt.Sprintf("(%v <= %v and %v >= %v)", self.Min, self.ColumnId, self.Max, self.ColumnId), nil

	default:
		return "", errors.New("Unknown test")

	}

}
