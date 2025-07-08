package repository

import (
	"cashier-api/helper/client"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/supabase-community/supabase-go"
)

func TestCategoryRepository(t *testing.T) {
	var supabaseClient *supabase.Client = client.CreateSupabaseClient()

	const (
		TENANT_ID int = 1
		STORE_ID  int = 1
	)

	t.Run("GetItemsByCategory", func(t *testing.T) {
		categoryRepositoryImpl := &CategoryRepositoryImpl{Client: supabaseClient}

		/*
			id (category id) id=1
			tenant_id=1
			limit=10
			page=1
		*/
		page := 1
		pagePerContent := 2
		categoryWithItemFromDB, err := categoryRepositoryImpl.GetItemsByCategory(1, TENANT_ID, pagePerContent, page-1)

		assert.Nil(t, err)
		assert.NotNil(t, categoryWithItemFromDB)
		assert.Equal(t, pagePerContent, len(categoryWithItemFromDB))
	})
}
