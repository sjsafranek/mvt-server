package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/sjsafranek/goutils/shell"
)

func UploadShapefile(shapefile, tablename, description string, srid int64) (string, error) {

	psql_connect := fmt.Sprintf(`PGPASSWORD=%v psql -d %v -U %v`, config.Database.Password, config.Database.Database, config.Database.Username)
	import_shapefile := fmt.Sprintf(`shp2pgsql -I "%v" "%v" | %v`, shapefile, tablename, psql_connect)
	create_layer := fmt.Sprintf(`%v -c "
        INSERT INTO layers (layer_name, description, srid) VALUES ('%v', '%v', %v)
    "`, psql_connect, strings.ToLower(tablename), description, srid)

	// bash script contents
	script := fmt.Sprintf(`
#!/bin/bash

%v
%v
	`, import_shapefile, create_layer)

	logger.Info(import_shapefile)
	logger.Info(create_layer)

	// write to bash script
	fh, err := ioutil.TempFile("", "mvt_upload.*.sh")
	if nil != err {
		return "", err
	}
	fmt.Fprintf(fh, script)
	fh.Close()
	defer os.Remove(fh.Name())

	// execute bash script
	return shell.RunScript("/bin/sh", fh.Name())
}
