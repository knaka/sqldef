// Integration test of mysqldef command.
//
// Test requirement:
//   - go command
//   - `mysql -uroot` must succeed
package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMysqldefCreateTable(t *testing.T) {
	resetTestDatabase()

	createTable1 := "CREATE TABLE users (\n" +
		"  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,\n" +
		"  name varchar(40) DEFAULT NULL,\n" +
		"  created_at datetime NOT NULL\n" +
		");"
	createTable2 := "CREATE TABLE bigdata (\n" +
		"  data bigint\n" +
		");"

	writeFile("schema.sql", createTable1+"\n"+createTable2)
	result := assertedExecute(t, "mysqldef", "-uroot", "mysqldef_test", "--file", "schema.sql")
	assertEquals(t, result, "Run: '"+createTable1+"'\n"+"Run: '"+createTable2+"'\n")

	writeFile("schema.sql", createTable1)
	result = assertedExecute(t, "mysqldef", "-uroot", "mysqldef_test", "--file", "schema.sql")
	assertEquals(t, result, "Run: 'DROP TABLE bigdata;'\n")
}

func TestMysqldefAddColumn(t *testing.T) {
	resetTestDatabase()

	createTable := "CREATE TABLE users (\n" +
		"  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,\n" +
		"  name varchar(40) DEFAULT NULL\n" +
		");"
	writeFile("schema.sql", createTable)
	result := assertedExecute(t, "mysqldef", "-uroot", "mysqldef_test", "--file", "schema.sql")
	assertEquals(t, result, "Run: '"+createTable+"'\n")

	createTable = "CREATE TABLE users (\n" +
		"  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,\n" +
		"  name varchar(40) DEFAULT NULL,\n" +
		"  created_at datetime NOT NULL\n" +
		");"
	writeFile("schema.sql", createTable)
	result = assertedExecute(t, "mysqldef", "-uroot", "mysqldef_test", "--file", "schema.sql")
	assertEquals(t, result, "Run: 'ALTER TABLE users ADD COLUMN created_at datetime NOT NULL ;'\n")

	createTable = "CREATE TABLE users (\n" +
		"  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,\n" +
		"  created_at datetime NOT NULL\n" +
		");"
	writeFile("schema.sql", createTable)
	result = assertedExecute(t, "mysqldef", "-uroot", "mysqldef_test", "--file", "schema.sql")
	assertEquals(t, result, "Run: 'ALTER TABLE users DROP COLUMN name;'\n")
}

func TestMysqldefAddIndex(t *testing.T) {
	resetTestDatabase()

	createTable := "CREATE TABLE users (\n" +
		"  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,\n" +
		"  name varchar(40) DEFAULT NULL,\n" +
		"  created_at datetime NOT NULL\n" +
		");"
	writeFile("schema.sql", createTable)
	assertedExecute(t, "mysqldef", "-uroot", "mysqldef_test", "--file", "schema.sql")

	alterTable := "ALTER TABLE users ADD INDEX index_name(name);"
	writeFile("schema.sql", createTable+"\n"+alterTable)
	result := assertedExecute(t, "mysqldef", "-uroot", "mysqldef_test", "--file", "schema.sql")
	assertEquals(t, result, "Run: '"+alterTable+"'\n")

	writeFile("schema.sql", createTable)
	result = assertedExecute(t, "mysqldef", "-uroot", "mysqldef_test", "--file", "schema.sql")
	assertEquals(t, result, "Run: 'ALTER TABLE users DROP INDEX index_name;'\n")
}

func TestMysqldefCreateTableKey(t *testing.T) {
	t.Skip() // Nothing is modified, for now.
	resetTestDatabase()

	createTable := "CREATE TABLE users (\n" +
		"  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,\n" +
		"  name varchar(40) DEFAULT NULL,\n" +
		"  created_at datetime NOT NULL\n" +
		");"
	writeFile("schema.sql", createTable)
	assertedExecute(t, "mysqldef", "-uroot", "mysqldef_test", "--file", "schema.sql")

	createTable = "CREATE TABLE users (\n" +
		"  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,\n" +
		"  name varchar(40) DEFAULT NULL,\n" +
		"  created_at datetime NOT NULL,\n" +
		"  KEY index_name(name)\n" +
		");"
	writeFile("schema.sql", createTable)
	result := assertedExecute(t, "mysqldef", "-uroot", "mysqldef_test", "--file", "schema.sql")
	assertEquals(t, result, "Run: 'ALTER TABLE users ADD INDEX index_name(name);'\n")
}

func TestMysqldefCreateTableSyntaxError(t *testing.T) {
	t.Skip() // invalid memory address or nil pointer dereference
	resetTestDatabase()

	createTable := "CREATE TABLE users (\n" +
		"  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,\n" +
		"  name varchar(40) DEFAULT NULL,\n" +
		"  created_at datetime NOT NULL,\n" +
		");"
	writeFile("schema.sql", createTable)
	assertedExecute(t, "mysqldef", "-uroot", "mysqldef_test", "--file", "schema.sql")
}

func TestMysqldefDryRun(t *testing.T) {
	resetTestDatabase()
	writeFile("schema.sql", `
	  CREATE TABLE users (
	    name varchar(40),
	    created_at datetime NOT NULL
	  );
	`)

	dryRun := assertedExecute(t, "mysqldef", "-uroot", "mysqldef_test", "--dry-run", "--file", "schema.sql")
	apply := assertedExecute(t, "mysqldef", "-uroot", "mysqldef_test", "--file", "schema.sql")
	assertEquals(t, dryRun, "--- dry run ---\n"+apply)
}

func TestMysqldefExport(t *testing.T) {
	resetTestDatabase()
	out := assertedExecute(t, "mysqldef", "-uroot", "mysqldef_test", "--export")
	assertEquals(t, out, "-- No table exists\n")

	mustExecute("mysql", "-uroot", "mysqldef_test", "-e", `
	  CREATE TABLE users (
	    name varchar(40),
	    created_at datetime NOT NULL
	  );
	`)
	out = assertedExecute(t, "mysqldef", "-uroot", "mysqldef_test", "--export")
	assertEquals(t, out,
		"CREATE TABLE `users` (\n"+
			"  `name` varchar(40) DEFAULT NULL,\n"+
			"  `created_at` datetime NOT NULL\n"+
			") ENGINE=InnoDB DEFAULT CHARSET=latin1;\n",
	)
}

func TestMysqldefHelp(t *testing.T) {
	_, err := execute("mysqldef", "--help")
	if err != nil {
		t.Errorf("failed to run --help: %s", err)
	}

	out, err := execute("mysqldef")
	if err == nil {
		t.Errorf("no database must be error, but successfully got: %s", out)
	}
}

func TestMain(m *testing.M) {
	resetTestDatabase()
	mustExecute("go", "build")
	status := m.Run()
	os.Exit(status)
}

func mustExecute(command string, args ...string) string {
	out, err := execute(command, args...)
	if err != nil {
		log.Printf("command: '%s %s'", command, strings.Join(args, " "))
		log.Printf("out: '%s'", out)
		log.Fatal(err)
	}
	return out
}

func assertedExecute(t *testing.T, command string, args ...string) string {
	out, err := execute(command, args...)
	if err != nil {
		t.Errorf("failed to execute '%s %s' (error: '%s'): `%s`", command, strings.Join(args, " "), err, out)
	}
	return out
}

func assertEquals(t *testing.T, actual string, expected string) {
	if expected != actual {
		t.Errorf("expected `%s` but got `%s`", expected, actual)
	}
}

func execute(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func resetTestDatabase() {
	mustExecute("mysql", "-uroot", "-e", "DROP DATABASE IF EXISTS mysqldef_test;")
	mustExecute("mysql", "-uroot", "-e", "CREATE DATABASE mysqldef_test;")
}

func writeFile(path string, content string) {
	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	file.Write(([]byte)(content))
}