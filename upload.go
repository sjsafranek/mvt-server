package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/sjsafranek/goutils/shell"
)

func UploadShapefile(shapefile, tablename, description string, srid int64) (string, error) {

	psql_connect := fmt.Sprintf(`PGPASSWORD=%v psql -d %v -U %v -h %v -p %v`, config.Database.Password, config.Database.Database, config.Database.Username, config.Database.Host, config.Database.Port)
	// import_shapefile := fmt.Sprintf(`shp2pgsql -I "%v" "%v" | %v`, shapefile, tablename, psql_connect)
	import_shapefile := fmt.Sprintf(`shp2pgsql -D -I "%v" "%v"`, shapefile, tablename)
	create_layer := fmt.Sprintf(`%v -c "
        INSERT INTO layers (layer_name, description, srid) VALUES ('%v', '%v', %v)
    "`, psql_connect, strings.ToLower(tablename), description, srid)

	// TODO use this -->
	// shp2pgsql -D -c $shpfile $table_name > tmp.dump
	// psql -d inrix -v ON_ERROR_STOP=1 --echo-errors -f tmp.dump

	// bash script contents
	script := fmt.Sprintf(`
#!/bin/bash
set -xe

%v > tmp.sql
result=$(%v -f tmp.sql)

[[ ! -z "$result" ]] && echo "failed to import file" && exit 1

%v
	`, import_shapefile, psql_connect, create_layer)

	// write to bash script
	fh, err := ioutil.TempFile("", "mvt_upload.*.sh")
	if nil != err {
		return "", err
	}
	fmt.Fprintf(fh, script)
	fh.Close()
	defer os.Remove(fh.Name())

	fmt.Println(script)

	// execute bash script
	return shell.RunScript("/bin/sh", fh.Name())
}

/*

{"method":"upload","file_path": "/home/stefan/DB4IoT/DS_10283_2597/GB_Postcodes/PostalSector.shp", "layer_name":"gb_postcodes_postalsector", "description":"gb_postcodes_postalsector", "srid": 27700}



{"method":"upload","file_path": "/home/stefan/DB4IoT/DS_10283_2597/GB_Postcodes/PostalSector2.shp", "layer_name":"gb_postcodes_postalsector", "description":"gb_postcodes_postalsector", "srid": 27700}

*/
