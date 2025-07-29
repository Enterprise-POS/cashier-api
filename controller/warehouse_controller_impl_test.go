package controller

import (
	"fmt"
	"os"
	"testing"
)

func TestWarehouseControllerImpl(t *testing.T) {
	fmt.Println(os.Getenv("MODE"))
}
