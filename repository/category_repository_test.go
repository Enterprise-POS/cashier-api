package repository

import (
	"fmt"
	"os"
	"testing"
)

func TestCategoryRepository(t *testing.T) {
	// var supabaseClient *supabase.Client = client.CreateSupabaseClient()

	const (
		TENANT_ID int = 1
		STORE_ID  int = 1
	)

	t.Run("GetItemsByCategory", func(t *testing.T) {
		// categoryRepositoryImpl := &CategoryRepositoryImpl{Client: supabaseClient}

		/*
			id (category id) id=1
			tenant_id=1
			limit=10
			page=1
		*/
		// categoryWithItemFromDB, count, err := categoryRepositoryImpl.GetItemsByCategory(1, TENANT_ID, 10, 0, false)
		// assert.Nil(t, err)
		// fmt.Println(count)
		// fmt.Println(categoryWithItemFromDB)

		// fmt.Println(categoryWithItemFromDB[0])
		env := os.Getenv("MODE")
		fmt.Println(env)
	})
}
