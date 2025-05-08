package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/jcelliott/lumber"
)

const Version = "1.0.0"

type (
	Logger interface {
		Fatal(string, ...interface{})
		Error(string, ...interface{})
		Warn(string, ...interface{})
		Info(string, ...interface{})
		Debug(string, ...interface{})
		Trace(string, ...interface{})
	}

	Driver struct {
		mutex   sync.Mutex
		mutexes map[string]*sync.Mutex
		dir     string
		log     Logger
	}
)

type Options struct {
	Logger
}

// ---------------------------NEW------------------------------------

func New(dir string, options *Options) (*Driver, error) {
	dir = filepath.Clean(dir)

	opts := Options{}

	if options != nil {
		opts = *options
	}

	if opts.Logger == nil {
		opts.Logger = lumber.NewConsoleLogger((lumber.INFO))
	}

	driver := Driver{
		dir:     dir,
		mutexes: make(map[string]*sync.Mutex),
		log:     opts.Logger,
	}

	// if DB already exists
	if _, err := os.Stat(dir); err == nil {
		opts.Logger.Debug("using '%s' (DB already exists)\n", dir)
		return &driver, nil
	}

	opts.Logger.Debug("Creating DB at '%s'...\n", dir)
	return &driver, os.MkdirAll(dir, 0755)
}

// ---------------------------WRITE------------------------------------

func (d *Driver) Write(collection, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("missing collection - no place to save the record")
	}

	if resource == "" {
		return fmt.Errorf("missing resource - unable to save record")
	}

	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, collection)
	fnlPath := filepath.Join(dir, resource+".json")
	tmpPath := fnlPath + ".tmp"

	// If path is already a directory, MkdirAll does nothing and returns nil.
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// marshal and indent the incoming data in Write fn.
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}

	b = append(b, byte('\n'))

	// writes data to the named file, creating it if necessary.
	if err := os.WriteFile(tmpPath, b, 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, fnlPath)
}

// ---------------------------READ------------------------------------

func (d *Driver) Read(collection, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("missing collection - no place to read the record")
	}

	if resource == "" {
		return fmt.Errorf("missing resource - unable to read record (no name)")
	}

	record := filepath.Join(d.dir, collection, resource)

	if _, err := stat(record); err != nil {
		return err
	}

	b, err := os.ReadFile(record + ".json")
	if err != nil {
		return err
	}

	return json.Unmarshal(b, &v)
}

func (d *Driver) ReadAll(collection string) ([]string, error) {
	if collection == "" {
		return nil, fmt.Errorf("missing collection - unable to read")
	}

	dir := filepath.Join(d.dir, collection)

	if _, err := stat(dir); err != nil {
		return nil, err
	}

	files, _ := os.ReadDir(dir)

	var records []string

	for _, file := range files {
		b, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}

		records = append(records, string(b))
	}

	return records, nil
}

func (d *Driver) Delete(collection, resource string) error {
	path := filepath.Join(collection, resource)
	mutex := d.getOrCreateMutex(collection)

	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, path)

	switch fi, err := stat(dir); {
	case fi == nil, err != nil:
		return fmt.Errorf("unable to find the file")
	case fi.Mode().IsDir():
		return os.RemoveAll(dir)
	case fi.Mode().IsRegular():
		return os.RemoveAll(dir + ".json")
	}

	return nil
}

// ---------------------------MUTEX------------------------------------

func (d *Driver) getOrCreateMutex(collection string) *sync.Mutex {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	m, ok := d.mutexes[collection]

	if !ok {
		m = &sync.Mutex{}
		d.mutexes[collection] = m
	}

	return m
}

func stat(path string) (fi os.FileInfo, err error) {
	if fi, err = os.Stat(path); os.IsNotExist(err) {
		fi, err = os.Stat(path + ".json")
	}

	return
}

type Address struct {
	City    string
	State   string
	Country string
	Pincode json.Number
}

type User struct {
	Name    string
	Age     json.Number
	Contact string
	Company string
	Address Address
}

func main() {
	dir := "./"

	db, err := New(dir, nil)
	if err != nil {
		fmt.Println("Error", err)
	}

	employees := []User{
		{"John", "23", "23213432", "My tech", Address{"bang", "Kt", "IN", "1234"}},
		{"Doe", "24", "23213432", "Your tech", Address{"hyd", "Kt", "IN", "1324"}},
		{"Hero", "25", "23213432", "In tech", Address{"del", "del", "IN", "6789"}},
		{"JSingh", "69", "2345678", "FU tech", Address{"BIJ", "pu", "IN", "8765"}},
	}

	for _, value := range employees {
		db.Write("users", value.Name, User{
			Name:    value.Name,
			Age:     value.Age,
			Contact: value.Contact,
			Company: value.Company,
			Address: value.Address,
		})
	}

	records, err := db.ReadAll("users")
	if err != nil {
		fmt.Println("Error", err)
	}

	fmt.Println(records)

	allusers := []User{}
	for _, f := range records {
		employeeFound := User{}

		if err := json.Unmarshal([]byte(f), &employeeFound); err != nil {
			fmt.Println("Error", err)
		}
		allusers = append(allusers, employeeFound)
	}
	fmt.Println((allusers))

	// if err := db.Delete("user", "john"); err != nil {
	// 	fmt.Println("Error", err)
	// }

	// if err := db.Delete("user", ""); err != nil {
	// 	fmt.Println("Error", err)
	// }
}
