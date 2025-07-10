package repository

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/supabase-community/supabase-go"
)

func TestCategoryRepository(t *testing.T) {
	var supabaseClient *supabase.Client = client.CreateSupabaseClient()

	const (
		CategoryTable string = "category"
		TenantId      int    = 1
		StoreId       int    = 1
	)

	t.Run("CreateCategoryRepositoryImplByNewFn", func(t *testing.T) {
		categoryRepositoryImpl := NewCategoryRepositoryImpl(supabaseClient)
		assert.NotNil(t, categoryRepositoryImpl)
		assert.NotNil(t, categoryRepositoryImpl.Client)
	})

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
			categoryWithItemFromDB, err := categoryRepositoryImpl.GetItemsByCategoryId(1, TenantId, pagePerContent, page-1)

			assert.Nil(t, err)
			assert.NotNil(t, categoryWithItemFromDB)
			assert.Equal(t, pagePerContent, len(categoryWithItemFromDB))
		})
		t.Run("NotExistCategoryId", func(t *testing.T) {
			page := 1
			pagePerContent := 2
			categoryWithItemsFromDB, err := categoryRepositoryImpl.GetItemsByCategoryId(0, TenantId, pagePerContent, page-1)
			assert.Nil(t, err)
			assert.Equal(t, 0, len(categoryWithItemsFromDB))
		})
		t.Run("NotExistIdAndOverflow", func(t *testing.T) {
			page := 100 // overflow
			pagePerContent := 2
			categoryWithItemsFromDB, err := categoryRepositoryImpl.GetItemsByCategoryId(0, TenantId, pagePerContent, page-1)
			assert.Nil(t, err)
			assert.Equal(t, 0, len(categoryWithItemsFromDB))
		})
		t.Run("NotExistTenantId", func(t *testing.T) {
			page := 1
			pagePerContent := 2
			categoryWithItemsFromDB, err := categoryRepositoryImpl.GetItemsByCategoryId(0, TenantId, pagePerContent, page-1)
			assert.Nil(t, err)
			assert.Equal(t, 0, len(categoryWithItemsFromDB))
		})
	})

	t.Run("GetCategoryWithItems", func(t *testing.T) {
		categoryRepositoryImpl := &CategoryRepositoryImpl{Client: supabaseClient}

		t.Run("NormalGet", func(t *testing.T) {
			page := 1
			pagePerContent := 3
			categoryWithItemFromDB, count, err := categoryRepositoryImpl.GetCategoryWithItems(TenantId, page-1, pagePerContent, true)
			assert.Nil(t, err)
			assert.NotEqual(t, 0, count)
			assert.NotNil(t, categoryWithItemFromDB)

			// Ignoring count
			categoryWithItemFromDB, count, err = categoryRepositoryImpl.GetCategoryWithItems(TenantId, page-1, pagePerContent, false)
			assert.Nil(t, err)
			assert.Equal(t, 0, count)
			assert.NotNil(t, categoryWithItemFromDB)
		})

		t.Run("Overflow", func(t *testing.T) {
			page := 2 // overflow
			pagePerContent := 100
			categoryWithItemFromDB, count, err := categoryRepositoryImpl.GetCategoryWithItems(TenantId, page-1, pagePerContent, true)
			assert.Nil(t, err)
			assert.NotEqual(t, 0, count)
			assert.NotNil(t, categoryWithItemFromDB)

			// Ignoring count
			categoryWithItemFromDB, count, err = categoryRepositoryImpl.GetCategoryWithItems(TenantId, page-1, pagePerContent, false)
			assert.Nil(t, err)
			assert.Equal(t, 0, count)
			assert.NotNil(t, categoryWithItemFromDB)
		})
	})

	t.Run("Get", func(t *testing.T) {
		categoryRepositoryImpl := &CategoryRepositoryImpl{Client: supabaseClient}

		t.Run("NormalGetAll", func(t *testing.T) {
			page := 1
			pagePerContent := 2
			categories, count, err := categoryRepositoryImpl.Get(TenantId, page-1, pagePerContent)
			assert.Nil(t, err)
			assert.NotEqual(t, 0, count)
			assert.NotNil(t, categories)
			assert.Greater(t, len(categories), 0)
		})

		t.Run("Overflow", func(t *testing.T) {
			// No error because correct page
			page := 1
			pagePerContent := 999
			categories, count, err := categoryRepositoryImpl.Get(TenantId, page-1, pagePerContent)
			assert.Greater(t, count, 0)
			assert.Nil(t, err)
			assert.NotNil(t, categories)

			// Error because page 2 doesn't even exist
			page = 2
			pagePerContent = 999
			categories, count, err = categoryRepositoryImpl.Get(TenantId, page-1, pagePerContent)

			assert.Equal(t, 0, count)
			assert.NotNil(t, err)
			assert.Equal(t, "(PGRST103) Requested range not satisfiable", err.Error())
			assert.Nil(t, categories)
		})
	})

	t.Run("Create", func(t *testing.T) {
		categoryRepositoryImpl := NewCategoryRepositoryImpl(supabaseClient)

		t.Run("CreateOne", func(t *testing.T) {
			dummyData := &model.Category{
				TenantId:     TenantId,
				CategoryName: "Test_CategoryRepositoryImpl_Create_CreateOne 1",
			}
			createdDummyCategoryFromDB, err := categoryRepositoryImpl.Create(TenantId, []*model.Category{dummyData})
			assert.Nil(t, err)
			assert.NotEqual(t, 0, createdDummyCategoryFromDB[0].Id)
			assert.NotEqual(t, 0, len(createdDummyCategoryFromDB))
			assert.Equal(t, dummyData.CategoryName, createdDummyCategoryFromDB[0].CategoryName)

			// Check using Get method if the data really placed in DB
			var actualCategory *model.Category
			count, err := supabaseClient.From(CategoryTable).Select("*", "", false).Eq("id", strconv.Itoa(createdDummyCategoryFromDB[0].Id)).Single().ExecuteTo(&actualCategory)
			require.Nil(t, err)
			assert.Equal(t, createdDummyCategoryFromDB[0].Id, actualCategory.Id)
			assert.Equal(t, createdDummyCategoryFromDB[0].CategoryName, actualCategory.CategoryName)

			// Clean up
			_, count, err = supabaseClient.From(CategoryTable).Delete("", "exact").Eq("category_name", dummyData.CategoryName).Execute()
			require.Nil(t, err, "Data test persist")
			require.Equal(t, 1, int(count))
		})
	})
}
