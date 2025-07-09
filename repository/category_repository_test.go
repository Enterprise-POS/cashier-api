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

		t.Run("NormalGet", func(t *testing.T) {
			/*
				id (category id) id=1
				tenant_id=1
				limit=10
				page=1
			*/
			page := 1
			pagePerContent := 2
			categoryWithItemFromDB, err := categoryRepositoryImpl.GetItemsByCategoryId(1, TENANT_ID, pagePerContent, page-1)

			assert.Nil(t, err)
			assert.NotNil(t, categoryWithItemFromDB)
			assert.Equal(t, pagePerContent, len(categoryWithItemFromDB))
		})
		t.Run("NotExistCategoryId", func(t *testing.T) {
			page := 1
			pagePerContent := 2
			categoryWithItemsFromDB, err := categoryRepositoryImpl.GetItemsByCategoryId(0, TENANT_ID, pagePerContent, page-1)
			assert.Nil(t, err)
			assert.Equal(t, 0, len(categoryWithItemsFromDB))
		})
		t.Run("NotExistIdAndOverflow", func(t *testing.T) {
			page := 100 // overflow
			pagePerContent := 2
			categoryWithItemsFromDB, err := categoryRepositoryImpl.GetItemsByCategoryId(0, TENANT_ID, pagePerContent, page-1)
			assert.Nil(t, err)
			assert.Equal(t, 0, len(categoryWithItemsFromDB))
		})
		t.Run("NotExistTenantId", func(t *testing.T) {
			page := 1
			pagePerContent := 2
			categoryWithItemsFromDB, err := categoryRepositoryImpl.GetItemsByCategoryId(0, TENANT_ID, pagePerContent, page-1)
			assert.Nil(t, err)
			assert.Equal(t, 0, len(categoryWithItemsFromDB))
		})
	})

	t.Run("GetCategoryWithItems", func(t *testing.T) {
		categoryRepositoryImpl := &CategoryRepositoryImpl{Client: supabaseClient}

		t.Run("NormalGet", func(t *testing.T) {
			page := 1
			pagePerContent := 3
			categoryWithItemFromDB, count, err := categoryRepositoryImpl.GetCategoryWithItems(TENANT_ID, page-1, pagePerContent, true)
			assert.Nil(t, err)
			assert.NotEqual(t, 0, count)
			assert.NotNil(t, categoryWithItemFromDB)
		})

		t.Run("Overflow", func(t *testing.T) {
			page := 2 // overflow
			pagePerContent := 100
			categoryWithItemFromDB, count, err := categoryRepositoryImpl.GetCategoryWithItems(TENANT_ID, page-1, pagePerContent, true)
			assert.Nil(t, err)
			assert.NotEqual(t, 0, count)
			assert.NotNil(t, categoryWithItemFromDB)
		})
	})
}
