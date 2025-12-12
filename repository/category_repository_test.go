package repository

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"fmt"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	"github.com/supabase-community/supabase-go"
)

func TestCategoryRepository(t *testing.T) {
	var supabaseClient *supabase.Client = client.CreateSupabaseClient()

	const (
		CategoryTable  string = "category"
		WarehouseTable string = "warehouse"
		TenantId       int    = 1
		StoreId        int    = 1
	)

	t.Run("CreateCategoryRepositoryImplByNewFn", func(t *testing.T) {
		categoryRepositoryImpl := NewCategoryRepositoryImpl(supabaseClient)
		assert.NotNil(t, categoryRepositoryImpl)
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
			categoryWithItemFromDB, count, err := categoryRepositoryImpl.GetItemsByCategoryId(1, TenantId, pagePerContent, page-1)

			assert.Nil(t, err)
			assert.NotNil(t, categoryWithItemFromDB)
			assert.Equal(t, pagePerContent, len(categoryWithItemFromDB))
			assert.NotEqual(t, 0, count)
		})
		t.Run("NotExistCategoryId", func(t *testing.T) {
			page := 1
			pagePerContent := 2
			categoryWithItemsFromDB, count, err := categoryRepositoryImpl.GetItemsByCategoryId(0, TenantId, pagePerContent, page-1)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(categoryWithItemsFromDB))
			assert.Equal(t, 0, count)
		})
		t.Run("NotExistIdAndOverflow", func(t *testing.T) {
			page := 100 // overflow
			pagePerContent := 2
			categoryWithItemsFromDB, count, err := categoryRepositoryImpl.GetItemsByCategoryId(0, TenantId, pagePerContent, page-1)
			assert.Nil(t, err)
			assert.Equal(t, 0, len(categoryWithItemsFromDB))
			assert.Equal(t, 0, count)
		})
		t.Run("NotExistTenantId", func(t *testing.T) {
			page := 1
			pagePerContent := 2
			categoryWithItemsFromDB, count, err := categoryRepositoryImpl.GetItemsByCategoryId(0, TenantId, pagePerContent, page-1)
			assert.Nil(t, err)
			assert.Equal(t, 0, len(categoryWithItemsFromDB))
			assert.Equal(t, 0, count)
		})
	})

	t.Run("GetCategoryWithItems", func(t *testing.T) {
		categoryRepositoryImpl := &CategoryRepositoryImpl{Client: supabaseClient}

		t.Run("NormalGet", func(t *testing.T) {
			page := 1
			pagePerContent := 3
			categoryWithItemFromDB, count, err := categoryRepositoryImpl.GetCategoryWithItems(TenantId, page-1, pagePerContent)
			assert.Nil(t, err)
			assert.NotEqual(t, 0, count)
			assert.NotNil(t, categoryWithItemFromDB)
		})

		t.Run("Overflow", func(t *testing.T) {
			page := 2 // overflow
			pagePerContent := 100
			categoryWithItemFromDB, count, err := categoryRepositoryImpl.GetCategoryWithItems(TenantId, page-1, pagePerContent)
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
			categories, count, err := categoryRepositoryImpl.Get(TenantId, page-1, pagePerContent, "")
			assert.Nil(t, err)
			assert.NotEqual(t, 0, count)
			assert.NotNil(t, categories)
			assert.Greater(t, len(categories), 0)
		})

		t.Run("Overflow", func(t *testing.T) {
			// No error because correct page
			page := 1
			pagePerContent := 999
			categories, count, err := categoryRepositoryImpl.Get(TenantId, page-1, pagePerContent, "")
			assert.Greater(t, count, 0)
			assert.Nil(t, err)
			assert.NotNil(t, categories)

			// Error because page 2 doesn't even exist
			page = 2
			pagePerContent = 999
			categories, count, err = categoryRepositoryImpl.Get(TenantId, page-1, pagePerContent, "")

			assert.Equal(t, 0, count)
			assert.NotNil(t, err)
			assert.Equal(t, "(PGRST103) Requested range not satisfiable", err.Error())
			assert.Nil(t, categories)
		})

		t.Run("GetByNameQuery", func(t *testing.T) {
			uniqueString := uuid.NewString()
			dummyCategories := []*model.Category{
				{
					// Id: ,
					CategoryName: "Test_GetByNameQuery_1_" + uniqueString,
					TenantId:     TenantId,
				},
				{
					// Id: ,
					CategoryName: "Test_GetByNameQuery_2_" + uniqueString,
					TenantId:     TenantId,
				},
			}

			// Manually insert without using another repository method
			var expectedDummyData []*model.Category
			_, err := supabaseClient.From(CategoryTable).
				Insert(dummyCategories, false, "", "representation", "").
				ExecuteTo(&expectedDummyData)
			require.NoError(t, err)
			require.NotNil(t, expectedDummyData)

			page := 1
			categories, count, err := categoryRepositoryImpl.Get(TenantId, 1, page-1, uniqueString)
			assert.NoError(t, err)
			assert.NotNil(t, categories)
			assert.Equal(t, len(dummyCategories), count)
			for i, category := range categories {
				assert.Equal(t, expectedDummyData[i].CategoryName, category.CategoryName)
				assert.Equal(t, expectedDummyData[i].TenantId, category.TenantId)
				assert.NotEqual(t, 0, category.Id)
				assert.NotNil(t, category.CreatedAt)
			}

			t.Cleanup(func() {
				_, _, err := supabaseClient.From(CategoryTable).
					Delete("", "").
					Eq("tenant_id", fmt.Sprint(TenantId)).
					In("id", []string{fmt.Sprint(expectedDummyData[0].Id), fmt.Sprint(expectedDummyData[1].Id)}).
					Execute()
				errorName := "TestCategoryRepository/Get/GetByNameQuery"
				require.NoErrorf(t, err, "If this fail then delete immediately, %s", errorName)
			})
		})
	})

	t.Run("Create", func(t *testing.T) {
		categoryRepositoryImpl := NewCategoryRepositoryImpl(supabaseClient)

		t.Run("CreateOne", func(t *testing.T) {
			dummyData := &model.Category{
				TenantId:     TenantId,
				CategoryName: "Test_CategoryRepositoryImpl_Create_CreateOne 1 " + uuid.NewString(),
			}
			createdDummyCategoryFromDB, err := categoryRepositoryImpl.Create(TenantId, []*model.Category{dummyData})
			assert.Nil(t, err)
			assert.NotEqual(t, 0, createdDummyCategoryFromDB[0].Id)
			assert.NotEqual(t, 0, len(createdDummyCategoryFromDB))
			assert.Equal(t, dummyData.CategoryName, createdDummyCategoryFromDB[0].CategoryName)

			// Check using Get method if the data really placed in DB
			var actualCategory *model.Category
			_, err = supabaseClient.From(CategoryTable).Select("*", "", false).Eq("id", strconv.Itoa(createdDummyCategoryFromDB[0].Id)).Single().ExecuteTo(&actualCategory)
			require.Nil(t, err)
			assert.Equal(t, createdDummyCategoryFromDB[0].Id, actualCategory.Id)
			assert.Equal(t, createdDummyCategoryFromDB[0].CategoryName, actualCategory.CategoryName)

			// Clean up
			_, count, err := supabaseClient.From(CategoryTable).Delete("", "exact").Eq("category_name", dummyData.CategoryName).Execute()
			require.Nil(t, err, "Data test persist")
			require.Equal(t, 1, int(count))
		})

		t.Run("CreateMultiple", func(t *testing.T) {
			dataDummies := []*model.Category{
				{
					TenantId:     TenantId,
					CategoryName: "Test_CategoryRepositoryImpl_Update_CreateMultiple 1",
				},
				{
					TenantId:     TenantId,
					CategoryName: "Test_CategoryRepositoryImpl_Update_CreateMultiple 2",
				},
				{
					TenantId:     TenantId,
					CategoryName: "Test_CategoryRepositoryImpl_Update_CreateMultiple 3",
				},
				{
					TenantId:     TenantId,
					CategoryName: "Test_CategoryRepositoryImpl_Update_CreateMultiple 4",
				},
				{
					TenantId:     TenantId,
					CategoryName: "Test_CategoryRepositoryImpl_Update_CreateMultiple 5",
				},
			}

			_createdDummyCategoryDB, err := categoryRepositoryImpl.Create(TenantId, dataDummies)
			require.Nil(t, err)
			require.Equal(t, 5, len(_createdDummyCategoryDB))

			for i, createdDummy := range _createdDummyCategoryDB {
				assert.Equal(t, dataDummies[i].CategoryName, createdDummy.CategoryName)
				assert.Equal(t, dataDummies[i].TenantId, createdDummy.TenantId)

				_, _, err = supabaseClient.From(CategoryTable).Delete("", "exact").Eq("category_name", createdDummy.CategoryName).Execute()
				require.Nil(t, err, "If this fail, immediately delete the test data; Create/CreatedMultiple")
			}
		})

		t.Run("CreateWithExistingId", func(t *testing.T) {
			dummyData := &model.Category{
				TenantId:     TenantId,
				CategoryName: "Test_CategoryRepositoryImpl_Update_CreateWithExistingId 1",
			}
			createdDummyCategoryFromDB, err := categoryRepositoryImpl.Create(TenantId, []*model.Category{dummyData})
			require.Nil(t, err)
			require.NotEqual(t, 0, createdDummyCategoryFromDB[0].Id)
			require.NotEqual(t, 0, len(createdDummyCategoryFromDB))
			require.Equal(t, dummyData.CategoryName, createdDummyCategoryFromDB[0].CategoryName)

			// Begin test; Assigning Id to Category is illegal insert,
			duplicateData := &model.Category{
				Id:           createdDummyCategoryFromDB[0].Id,
				TenantId:     TenantId,
				CategoryName: "Test_CategoryRepositoryImpl_Update_CreateMultiple 1",
			}
			duplicateDataFromDB, err := categoryRepositoryImpl.Create(TenantId, []*model.Category{duplicateData})
			assert.NotNil(t, err)
			assert.Nil(t, duplicateDataFromDB)
			assert.Equal(t, "(23505) duplicate key value violates unique constraint \"category_pkey\"", err.Error())

			// Clean up
			_, _, err = supabaseClient.From(CategoryTable).Delete("", "").Eq("category_name", dummyData.CategoryName).Execute()
			require.Nil(t, err, "If this fail, immediately delete the test data; Create/CreatedMultiple")
		})

		t.Run("CreateWithExactCategoryName", func(t *testing.T) {
			dummyData := &model.Category{
				TenantId:     TenantId,
				CategoryName: "Test_CategoryRepositoryImpl_Update_CreateWithExactCategoryName 1",
			}
			createdDummyCategoryFromDB, err := categoryRepositoryImpl.Create(TenantId, []*model.Category{dummyData})
			require.Nil(t, err)
			require.NotEqual(t, 0, createdDummyCategoryFromDB[0].Id)
			require.NotEqual(t, 0, len(createdDummyCategoryFromDB))
			require.Equal(t, dummyData.CategoryName, createdDummyCategoryFromDB[0].CategoryName)

			// Begin test; Exact category name is not allowed
			// example Fruits == Fruits
			// Lower case but the same mean is allowed
			// example Fruits != fruits
			duplicateData := &model.Category{
				TenantId:     TenantId,
				CategoryName: "Test_CategoryRepositoryImpl_Update_CreateWithExactCategoryName 1",
			}
			duplicateDataFromDB, err := categoryRepositoryImpl.Create(TenantId, []*model.Category{duplicateData})
			assert.NotNil(t, err)
			assert.Nil(t, duplicateDataFromDB)
			assert.Equal(t, "(23505) duplicate key value violates unique constraint \"unique_tenant_category_name\"", err.Error())

			// Clean up
			_, _, err = supabaseClient.From(CategoryTable).Delete("", "").Eq("category_name", dummyData.CategoryName).Execute()
			require.Nil(t, err, "If this fail, immediately delete the test data; Create/CreatedMultiple")
		})
	})

	t.Run("Register", func(t *testing.T) {
		// This is special insert, because category is many to many into warehouse
		categoryRepositoryImpl := NewCategoryRepositoryImpl(supabaseClient)
		warehouseRepositoryImpl := WarehouseRepositoryImpl{Client: supabaseClient}

		t.Run("NormalRegister", func(t *testing.T) {
			// Create warehouse item
			dummyItem := &model.Item{
				ItemName:  "Test_CategoryRepositoryImpl_Register_NormalRegister 1",
				Stocks:    10,
				TenantId:  TenantId,
				IsActive:  true,
				StockType: model.StockTypeTracked,
			}
			_dummyItemFromDB, err := warehouseRepositoryImpl.CreateItem([]*model.Item{dummyItem})
			require.Nil(t, err)
			require.Equal(t, 1, len(_dummyItemFromDB))

			// warehouse item
			dummyItemFromDB := _dummyItemFromDB[0]

			// Create the Category
			dummyCategory := &model.Category{
				CategoryName: "Test_CategoryRepositoryImpl_Register_NormalRegister 1 The Category",
				TenantId:     TenantId,
			}

			_createdDummyCategoryFromDB, err := categoryRepositoryImpl.Create(dummyCategory.TenantId, []*model.Category{dummyCategory})
			require.Nil(t, err)
			require.Equal(t, 1, len(_createdDummyCategoryFromDB))

			// category
			createdDummyCategoryFromDB := _createdDummyCategoryFromDB[0]

			dummyCategoryMtmWarehouse := []*model.CategoryMtmWarehouse{
				{
					CategoryId: createdDummyCategoryFromDB.Id,
					ItemId:     dummyItemFromDB.ItemId,
				},
			}

			// The test itself
			err = categoryRepositoryImpl.Register(dummyCategoryMtmWarehouse)
			assert.Nil(t, err)

			// Clean up
			_, _, err = supabaseClient.From("category_mtm_warehouse").Delete("", "").Eq("category_id", strconv.Itoa(createdDummyCategoryFromDB.Id)).Eq("item_id", strconv.Itoa(dummyItemFromDB.ItemId)).Execute()
			require.Nil(t, err, "If this fail, immediately delete the test data; Register/NormalRegister 1")
			_, _, err = supabaseClient.From(CategoryTable).Delete("", "").Eq("category_name", createdDummyCategoryFromDB.CategoryName).Execute()
			require.Nil(t, err, "If this fail, immediately delete the test data; Register/NormalRegister 2")
			_, _, err = supabaseClient.From(WarehouseTable).Delete("", "").Eq("item_name", dummyItemFromDB.ItemName).Execute()
			require.Nil(t, err, "If this fail, immediately delete the test data; Register/NormalRegister 3")
		})

		t.Run("DuplicateRegister", func(t *testing.T) {
			// Create warehouse item
			dummyItem := &model.Item{
				ItemName:  "Test_CategoryRepositoryImpl_Register_DuplicateRegister 1",
				Stocks:    10,
				StockType: model.StockTypeTracked,
				TenantId:  TenantId,
				IsActive:  true,
			}
			_dummyItemFromDB, err := warehouseRepositoryImpl.CreateItem([]*model.Item{dummyItem})
			require.Nil(t, err)
			require.Equal(t, 1, len(_dummyItemFromDB))

			// warehouse item
			dummyItemFromDB := _dummyItemFromDB[0]

			// Create the Category
			dummyCategory := &model.Category{
				CategoryName: "Test_CategoryRepositoryImpl_Register_DuplicateRegister 1 The Category",
				TenantId:     TenantId,
			}

			_createdDummyCategoryFromDB, err := categoryRepositoryImpl.Create(dummyCategory.TenantId, []*model.Category{dummyCategory})
			require.Nil(t, err)
			require.Equal(t, 1, len(_createdDummyCategoryFromDB))

			// category
			createdDummyCategoryFromDB := _createdDummyCategoryFromDB[0]

			dummyCategoryMtmWarehouse := []*model.CategoryMtmWarehouse{
				{
					CategoryId: createdDummyCategoryFromDB.Id,
					ItemId:     dummyItemFromDB.ItemId,
				},
			}

			// Test itself, repeat twice
			err = categoryRepositoryImpl.Register(dummyCategoryMtmWarehouse)
			assert.Nil(t, err)

			err = categoryRepositoryImpl.Register(dummyCategoryMtmWarehouse)
			assert.NotNil(t, err)
			assert.Equal(t, "(23505) duplicate key value violates unique constraint \"unique_category_mtm_warehouse_category_id_and_item_id\"", err.Error())

			// Clean up
			_, _, err = supabaseClient.From("category_mtm_warehouse").Delete("", "").Eq("category_id", strconv.Itoa(createdDummyCategoryFromDB.Id)).Eq("item_id", strconv.Itoa(dummyItemFromDB.ItemId)).Execute()
			require.Nil(t, err, "If this fail, immediately delete the test data; Register/DuplicateRegister 1")
			_, _, err = supabaseClient.From(CategoryTable).Delete("", "").Eq("category_name", createdDummyCategoryFromDB.CategoryName).Execute()
			require.Nil(t, err, "If this fail, immediately delete the test data; Register/DuplicateRegister 2")
			_, _, err = supabaseClient.From(WarehouseTable).Delete("", "").Eq("item_name", dummyItemFromDB.ItemName).Execute()
			require.Nil(t, err, "If this fail, immediately delete the test data; Register/DuplicateRegister 3")
		})
	})

	t.Run("Unregister", func(t *testing.T) {
		// This is special insert, because category is many to many into warehouse
		categoryRepositoryImpl := NewCategoryRepositoryImpl(supabaseClient)
		warehouseRepositoryImpl := WarehouseRepositoryImpl{Client: supabaseClient}

		t.Run("NormalUnregister", func(t *testing.T) {
			// START:
			// Create warehouse item
			dummyItem := &model.Item{
				ItemName:  "Test_CategoryRepositoryImpl_Register_NormalUnregister 1",
				Stocks:    10,
				StockType: model.StockTypeTracked,
				TenantId:  TenantId,
				IsActive:  true,
			}
			_dummyItemFromDB, err := warehouseRepositoryImpl.CreateItem([]*model.Item{dummyItem})
			require.Nil(t, err)
			require.Equal(t, 1, len(_dummyItemFromDB))

			// warehouse item
			dummyItemFromDB := _dummyItemFromDB[0]

			// Create the Category
			dummyCategory := &model.Category{
				CategoryName: "Test_CategoryRepositoryImpl_Register_NormalUnregister 1 The Category",
				TenantId:     TenantId,
			}

			_createdDummyCategoryFromDB, err := categoryRepositoryImpl.Create(dummyCategory.TenantId, []*model.Category{dummyCategory})
			require.Nil(t, err)
			require.Equal(t, 1, len(_createdDummyCategoryFromDB))

			// category
			createdDummyCategoryFromDB := _createdDummyCategoryFromDB[0]

			dummyCategoryMtmWarehouse := []*model.CategoryMtmWarehouse{
				{
					CategoryId: createdDummyCategoryFromDB.Id,
					ItemId:     dummyItemFromDB.ItemId,
				},
			}

			// END: Until here, is the same as the code above
			err = categoryRepositoryImpl.Register(dummyCategoryMtmWarehouse)
			assert.Nil(t, err)

			// The test itself
			tobeUnregisteredCategoryMtmWarehouse := &model.CategoryMtmWarehouse{
				CategoryId: createdDummyCategoryFromDB.Id,
				ItemId:     dummyItemFromDB.ItemId,
			}
			err = categoryRepositoryImpl.Unregister(tobeUnregisteredCategoryMtmWarehouse)
			assert.Nil(t, err)

			// Clean up
			// _, _, err = supabaseClient.From("category_mtm_warehouse").Delete("", "").Eq("category_id", strconv.Itoa(createdDummyCategoryFromDB.Id)).Eq("item_id", strconv.Itoa(dummyItemFromDB.ItemId)).Execute()
			// require.Nil(t, err, "If this fail, immediately delete the test data; Register/NormalUnregister 1")
			_, _, err = supabaseClient.From(CategoryTable).Delete("", "").Eq("category_name", createdDummyCategoryFromDB.CategoryName).Execute()
			require.Nil(t, err, "If this fail, immediately delete the test data; Register/NormalUnregister 2")
			_, _, err = supabaseClient.From(WarehouseTable).Delete("", "").Eq("item_name", dummyItemFromDB.ItemName).Execute()
			require.Nil(t, err, "If this fail, immediately delete the test data; Register/NormalUnregister 3")
		})

		t.Run("UnregisterThatUnregistered", func(t *testing.T) {
			tobeUnregisteredCategoryMtmWarehouse1 := &model.CategoryMtmWarehouse{
				CategoryId: 0,
				ItemId:     1,
			}
			tobeUnregisteredCategoryMtmWarehouse2 := &model.CategoryMtmWarehouse{
				CategoryId: 1,
				ItemId:     0,
			}

			err := categoryRepositoryImpl.Unregister(tobeUnregisteredCategoryMtmWarehouse1)
			require.NotNil(t, err, "If this fail, immediately check the the test data; there is a possibility data deleted !")
			require.Contains(t, err.Error(), "[WARN]")

			err = categoryRepositoryImpl.Unregister(tobeUnregisteredCategoryMtmWarehouse2)
			require.NotNil(t, err, "If this fail, immediately check the the test data; there is a possibility data deleted !")
			require.Contains(t, err.Error(), "[WARN]")
		})
	})

	t.Run("EditItemCategory", func(t *testing.T) {
		categoryRepositoryImpl := NewCategoryRepositoryImpl(supabaseClient)
		warehouseRepositoryImpl := NewWarehouseRepositoryImpl(supabaseClient)

		t.Run("NormalEditItemCategory", func(t *testing.T) {
			// START:
			// Create warehouse item
			dummyItem := &model.Item{
				ItemName:  "Test_CategoryRepositoryImpl_EditItemCategory_NormalEditItemCategory 1",
				Stocks:    10,
				StockType: model.StockTypeTracked,
				TenantId:  TenantId,
				IsActive:  true,
			}
			_dummyItemFromDB, err := warehouseRepositoryImpl.CreateItem([]*model.Item{dummyItem})
			require.Nil(t, err)
			require.Equal(t, 1, len(_dummyItemFromDB))

			// warehouse item
			dummyItemFromDB := _dummyItemFromDB[0]

			// Create the Category
			dummyCategories := []*model.Category{
				{
					CategoryName: "Test_CategoryRepositoryImpl_EditItemCategory_NormalEditItemCategory 1 The Category",
					TenantId:     TenantId,
				},
				{
					CategoryName: "Test_CategoryRepositoryImpl_EditItemCategory_NormalEditItemCategory 2 The Category",
					TenantId:     TenantId,
				},
			}

			_createdDummyCategoryFromDB, err := categoryRepositoryImpl.Create(TenantId, dummyCategories)
			require.Nil(t, err)
			require.Len(t, _createdDummyCategoryFromDB, len(dummyCategories))

			// category
			createdDummyCategoryFromDB1 := _createdDummyCategoryFromDB[0]
			createdDummyCategoryFromDB2 := _createdDummyCategoryFromDB[1]

			dummyCategoryMtmWarehouse := []*model.CategoryMtmWarehouse{
				{
					CategoryId: createdDummyCategoryFromDB1.Id,
					ItemId:     dummyItemFromDB.ItemId,
				},
			}

			err = categoryRepositoryImpl.Register(dummyCategoryMtmWarehouse)
			require.Nil(t, err)

			// The test itself
			tobeUpdateItemCategory := &model.CategoryMtmWarehouse{
				CategoryId: createdDummyCategoryFromDB2.Id, // Edited
				ItemId:     dummyItemFromDB.ItemId,
			}
			err = categoryRepositoryImpl.EditItemCategory(TenantId, tobeUpdateItemCategory)
			assert.NoError(t, err)

			// Delete the data
			// No need to unregister because when category deleted then all the data will be deleted

			// Clean up
			_, count, err := supabaseClient.From(CategoryTable).
				Delete("", "exact").
				In("id", []string{fmt.Sprint(createdDummyCategoryFromDB1.Id), fmt.Sprint(createdDummyCategoryFromDB2.Id)}).
				Execute()
			require.Nil(t, err, "If this fail, immediately delete the test data; EditItemCategory/NormalEditItemCategory 1")
			require.Equal(t, 2, int(count))
			_, _, err = supabaseClient.From(WarehouseTable).
				Delete("", "").
				Eq("item_name", dummyItemFromDB.ItemName).
				Eq("item_id", fmt.Sprint(dummyItemFromDB.ItemId)).
				Execute()
			require.Nil(t, err, "If this fail, immediately delete the test data; EditItemCategory/NormalEditItemCategory 2")
		})

		t.Run("NonExistenceItemAtWarehouse", func(t *testing.T) {
			// Expected to get this error message
			// [ERROR] Fatal error, current item from store never exist at warehouse
			tobeUpdateItemCategory := &model.CategoryMtmWarehouse{
				CategoryId: 1,        // Technically valid
				ItemId:     99999999, // Not available ItemId
			}
			err := categoryRepositoryImpl.EditItemCategory(TenantId, tobeUpdateItemCategory)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "[ERROR]")
		})

		t.Run("InvalidRequest", func(t *testing.T) {
			// User will act as unregister by sending categoryId: 0
			tobeUpdateItemCategory := &model.CategoryMtmWarehouse{
				CategoryId: 0, // Invalid input request
				ItemId:     1, // Technically valid
			}
			err := categoryRepositoryImpl.EditItemCategory(TenantId, tobeUpdateItemCategory)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "[ERROR]")
		})
	})

	t.Run("Update", func(t *testing.T) {
		categoryRepositoryImpl := NewCategoryRepositoryImpl(supabaseClient)

		t.Run("NormalUpdate", func(t *testing.T) {
			dummyData := &model.Category{
				TenantId:     TenantId,
				CategoryName: "Test_CategoryRepositoryImpl_Update_NormalUpdate 1",
			}
			_createdDummyCategoryDB, err := categoryRepositoryImpl.Create(TenantId, []*model.Category{dummyData})
			require.Nil(t, err)
			require.Equal(t, 1, len(_createdDummyCategoryDB))

			// Begin updating; only CategoryName could be updated
			createdDummyCategoryDB := _createdDummyCategoryDB[0]
			createdDummyCategoryDB.CategoryName = "Test_CategoryRepositoryImpl_Update_NormalUpdate 1 (UPDATED)"

			editedDummyCategoryDB, err := categoryRepositoryImpl.Update(createdDummyCategoryDB.TenantId, createdDummyCategoryDB.Id, createdDummyCategoryDB.CategoryName)
			assert.Nil(t, err)
			assert.NotNil(t, editedDummyCategoryDB)
			assert.Equal(t, createdDummyCategoryDB.Id, editedDummyCategoryDB.Id)
			assert.Equal(t, createdDummyCategoryDB.TenantId, editedDummyCategoryDB.TenantId)
			assert.Equal(t, createdDummyCategoryDB.CategoryName, editedDummyCategoryDB.CategoryName)
			assert.Equal(t, createdDummyCategoryDB.CreatedAt.UTC().Day(), editedDummyCategoryDB.CreatedAt.UTC().Day())

			// Clean up
			supabaseClient.From(CategoryTable).
				Delete("", "").
				Eq("tenant_id", strconv.Itoa(TenantId)).
				Eq("category_name", editedDummyCategoryDB.CategoryName).
				Execute()
		})

		t.Run("UpdateThatCategoryNotExist", func(t *testing.T) {
			editedDummyCategoryDB, err := categoryRepositoryImpl.Update(TenantId, 0, "Will not happen")
			assert.NotNil(t, err)
			assert.Nil(t, editedDummyCategoryDB)
			assert.Equal(t, "(PGRST116) JSON object requested, multiple (or no) rows returned", err.Error())
		})
	})

	t.Run("Delete", func(t *testing.T) {
		categoryRepositoryImpl := NewCategoryRepositoryImpl(supabaseClient)

		t.Run("NormalDelete", func(t *testing.T) {
			dummyData := &model.Category{
				TenantId:     TenantId,
				CategoryName: "Test_CategoryRepositoryImpl_Delete_NormalDelete 1",
			}
			_createdDummyCategoryDB, err := categoryRepositoryImpl.Create(TenantId, []*model.Category{dummyData})
			require.Nil(t, err)
			require.Equal(t, 1, len(_createdDummyCategoryDB))

			// The created category itself
			createdDummyCategoryDB := _createdDummyCategoryDB[0]

			err = categoryRepositoryImpl.Delete(createdDummyCategoryDB)
			assert.Nil(t, err)
		})

		t.Run("DeleteThatCategoryNotExist", func(t *testing.T) {
			dummyData := &model.Category{
				Id:           0,
				TenantId:     TenantId,
				CategoryName: "Test_CategoryRepositoryImpl_Delete_DeleteThatCategoryNotExist 1",
			}
			err := categoryRepositoryImpl.Delete(dummyData)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "[WARN]")
		})
	})
}
