package repository

import (
	"cashier-api/helper/client"
	"cashier-api/helper/query"
	"cashier-api/model"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
)

func TestCategoryRepository(t *testing.T) {
	var gormClient *gorm.DB = client.CreateGormClient()

	t.Run("CreateCategoryRepositoryImplByNewFn", func(t *testing.T) {
		tx := gormClient.Begin()
		defer tx.Rollback()

		categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)
		assert.NotNil(t, categoryRepositoryImpl)
	})

	t.Run("GetItemsByCategory", func(t *testing.T) {
		t.Run("NormalGet", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_getitemsbycategory_normal@example.com", "Category GetItemsByCategory Normal")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)
			warehouseRepositoryImpl := NewWarehouseRepositoryImpl(tx)

			dummyItems := []*model.Item{
				{ItemName: "Item 1 " + uuid.NewString(), Stocks: 10, TenantId: tenantId, IsActive: true, StockType: model.StockTypeTracked},
				{ItemName: "Item 2 " + uuid.NewString(), Stocks: 10, TenantId: tenantId, IsActive: true, StockType: model.StockTypeTracked},
			}
			createdItems, err := warehouseRepositoryImpl.CreateItem(dummyItems)
			require.NoError(t, err)
			require.Len(t, createdItems, 2)

			dummyCategory := &model.Category{TenantId: tenantId, CategoryName: "Category " + uuid.NewString()}
			createdCategories, err := categoryRepositoryImpl.Create(tenantId, []*model.Category{dummyCategory})
			require.NoError(t, err)
			createdCategory := createdCategories[0]

			var registrations []*model.CategoryMtmWarehouse
			for _, item := range createdItems {
				registrations = append(registrations, &model.CategoryMtmWarehouse{
					CategoryId: createdCategory.Id,
					ItemId:     item.ItemId,
				})
			}
			require.NoError(t, categoryRepositoryImpl.Register(registrations))

			page := 1
			pagePerContent := 2
			// Signature is (tenantId, categoryId, limit, page) — fixed order below.
			categoryWithItemFromDB, count, err := categoryRepositoryImpl.GetItemsByCategoryId(tenantId, createdCategory.Id, pagePerContent, page-1)

			assert.Nil(t, err)
			assert.NotNil(t, categoryWithItemFromDB)
			assert.Equal(t, pagePerContent, len(categoryWithItemFromDB))
			assert.NotEqual(t, 0, count)
		})

		t.Run("NotExistCategoryId", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_getitemsbycategory_notexistcat@example.com", "Category GetItemsByCategory NotExistCat")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)

			page := 1
			pagePerContent := 2
			categoryWithItemsFromDB, count, err := categoryRepositoryImpl.GetItemsByCategoryId(tenantId, 0, pagePerContent, page-1)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(categoryWithItemsFromDB))
			assert.Equal(t, 0, count)
		})

		t.Run("NotExistIdAndOverflow", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_getitemsbycategory_overflow@example.com", "Category GetItemsByCategory Overflow")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)

			page := 100 // overflow
			pagePerContent := 2
			categoryWithItemsFromDB, count, err := categoryRepositoryImpl.GetItemsByCategoryId(tenantId, 0, pagePerContent, page-1)
			assert.Nil(t, err)
			assert.Equal(t, 0, len(categoryWithItemsFromDB))
			assert.Equal(t, 0, count)
		})

		t.Run("NotExistTenantId", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_getitemsbycategory_notexisttenant@example.com", "Category GetItemsByCategory NotExistTenant")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)

			page := 1
			pagePerContent := 2
			// tenantId+999999 was never seeded, so it must not exist.
			categoryWithItemsFromDB, count, err := categoryRepositoryImpl.GetItemsByCategoryId(tenantId+999999, 0, pagePerContent, page-1)
			assert.Nil(t, err)
			assert.Equal(t, 0, len(categoryWithItemsFromDB))
			assert.Equal(t, 0, count)
		})
	})

	t.Run("GetCategoryWithItems", func(t *testing.T) {
		// seedCategoryWithItemsFixture seeds two categories, each with items
		// registered against the warehouse, for use across sub-tests below.
		seedCategoryWithItemsFixture := func(t *testing.T, tx *gorm.DB, tenantId int) (categoryA, categoryB *model.Category, itemA1, itemA2, itemB1 *model.Item) {
			t.Helper()

			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)
			warehouseRepositoryImpl := NewWarehouseRepositoryImpl(tx)

			dummyItemA1 := &model.Item{ItemName: "Apple_" + uuid.NewString(), Stocks: 10, BasePrice: 1000, TenantId: tenantId, IsActive: true, StockType: model.StockTypeTracked}
			dummyItemA2 := &model.Item{ItemName: "Banana_" + uuid.NewString(), Stocks: 10, BasePrice: 2000, TenantId: tenantId, IsActive: true, StockType: model.StockTypeTracked}
			dummyItemB1 := &model.Item{ItemName: "Cherry_" + uuid.NewString(), Stocks: 10, BasePrice: 3000, TenantId: tenantId, IsActive: true, StockType: model.StockTypeTracked}

			createdItems, err := warehouseRepositoryImpl.CreateItem([]*model.Item{dummyItemA1, dummyItemA2, dummyItemB1})
			require.NoError(t, err)
			require.Len(t, createdItems, 3)

			dummyCategoryA := &model.Category{TenantId: tenantId, CategoryName: "CategoryA_" + uuid.NewString()}
			dummyCategoryB := &model.Category{TenantId: tenantId, CategoryName: "CategoryB_" + uuid.NewString()}
			createdCategories, err := categoryRepositoryImpl.Create(tenantId, []*model.Category{dummyCategoryA, dummyCategoryB})
			require.NoError(t, err)
			require.Len(t, createdCategories, 2)

			require.NoError(t, categoryRepositoryImpl.Register([]*model.CategoryMtmWarehouse{
				{CategoryId: createdCategories[0].Id, ItemId: createdItems[0].ItemId},
				{CategoryId: createdCategories[0].Id, ItemId: createdItems[1].ItemId},
				{CategoryId: createdCategories[1].Id, ItemId: createdItems[2].ItemId},
			}))

			return createdCategories[0], createdCategories[1], createdItems[0], createdItems[1], createdItems[2]
		}

		t.Run("NormalGet", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_getcategorywithitems_normal@example.com", "Category GetCategoryWithItems Normal")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)
			seedCategoryWithItemsFixture(t, tx, tenantId)

			page := 1
			pagePerContent := 3
			categoryWithItemFromDB, count, err := categoryRepositoryImpl.GetCategoryWithItems(tenantId, page-1, pagePerContent, "", 0, []query.QueryFilter{})
			assert.Nil(t, err)
			assert.NotEqual(t, 0, count)
			assert.NotNil(t, categoryWithItemFromDB)
		})

		t.Run("Overflow", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_getcategorywithitems_overflow@example.com", "Category GetCategoryWithItems Overflow")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)
			seedCategoryWithItemsFixture(t, tx, tenantId)

			page := 1000 // overflow
			pagePerContent := 100
			categoryWithItemFromDB, count, err := categoryRepositoryImpl.GetCategoryWithItems(tenantId, page-1, pagePerContent, "", 0, []query.QueryFilter{})
			// Gorm treat overflow as nothing to return
			assert.NoError(t, err)
			assert.Equal(t, 0, count)
			assert.Empty(t, categoryWithItemFromDB)
		})

		t.Run("FilterByNameQuery", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_getcategorywithitems_namequery@example.com", "Category GetCategoryWithItems NameQuery")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)
			dbCategoryA, _, dbItemA1, _, _ := seedCategoryWithItemsFixture(t, tx, tenantId)

			page := 1
			pagePerContent := 10
			categoryWithItemFromDB, count, err := categoryRepositoryImpl.GetCategoryWithItems(tenantId, page-1, pagePerContent, dbItemA1.ItemName, 0, []query.QueryFilter{})
			assert.NoError(t, err)
			assert.Equal(t, 1, count)
			require.Len(t, categoryWithItemFromDB, 1)
			assert.Equal(t, dbItemA1.ItemId, categoryWithItemFromDB[0].ItemId)
			assert.Equal(t, dbCategoryA.Id, categoryWithItemFromDB[0].CategoryId)
		})

		t.Run("FilterByNameQueryNotFound", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_getcategorywithitems_namequerymiss@example.com", "Category GetCategoryWithItems NameQueryMiss")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)
			seedCategoryWithItemsFixture(t, tx, tenantId)

			page := 1
			pagePerContent := 10
			categoryWithItemFromDB, count, err := categoryRepositoryImpl.GetCategoryWithItems(tenantId, page-1, pagePerContent, "Nonexistent_"+uuid.NewString(), 0, []query.QueryFilter{})
			assert.NoError(t, err)
			assert.Equal(t, 0, count)
			assert.Empty(t, categoryWithItemFromDB)
		})

		t.Run("FilterByCategoryId", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_getcategorywithitems_categoryid@example.com", "Category GetCategoryWithItems CategoryId")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)
			dbCategoryA, _, _, _, _ := seedCategoryWithItemsFixture(t, tx, tenantId)

			page := 1
			pagePerContent := 10
			categoryWithItemFromDB, count, err := categoryRepositoryImpl.GetCategoryWithItems(tenantId, page-1, pagePerContent, "", dbCategoryA.Id, []query.QueryFilter{})
			assert.NoError(t, err)
			assert.Equal(t, 2, count)
			require.Len(t, categoryWithItemFromDB, 2)
			for _, result := range categoryWithItemFromDB {
				assert.Equal(t, dbCategoryA.Id, result.CategoryId)
			}
		})

		t.Run("OrderByFilterAscending", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_getcategorywithitems_orderasc@example.com", "Category GetCategoryWithItems OrderAsc")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)
			seedCategoryWithItemsFixture(t, tx, tenantId)

			page := 1
			pagePerContent := 10
			categoryWithItemFromDB, count, err := categoryRepositoryImpl.GetCategoryWithItems(
				tenantId, page-1, pagePerContent, "", 0,
				[]query.QueryFilter{{Column: "base_price", Ascending: true}},
			)
			assert.NoError(t, err)
			assert.NotEqual(t, 0, count)
			require.NotEmpty(t, categoryWithItemFromDB)
			for i := 1; i < len(categoryWithItemFromDB); i++ {
				assert.LessOrEqual(t, categoryWithItemFromDB[i-1].BasePrice, categoryWithItemFromDB[i].BasePrice)
			}
		})

		t.Run("OrderByFilterDescending", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_getcategorywithitems_orderdesc@example.com", "Category GetCategoryWithItems OrderDesc")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)
			seedCategoryWithItemsFixture(t, tx, tenantId)

			page := 1
			pagePerContent := 10
			categoryWithItemFromDB, count, err := categoryRepositoryImpl.GetCategoryWithItems(
				tenantId, page-1, pagePerContent, "", 0,
				[]query.QueryFilter{{Column: "base_price", Ascending: false}},
			)
			assert.NoError(t, err)
			assert.NotEqual(t, 0, count)
			require.NotEmpty(t, categoryWithItemFromDB)
			for i := 1; i < len(categoryWithItemFromDB); i++ {
				assert.GreaterOrEqual(t, categoryWithItemFromDB[i-1].BasePrice, categoryWithItemFromDB[i].BasePrice)
			}
		})

		t.Run("EmptyFilterColumnReturnsError", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_getcategorywithitems_emptyfilter@example.com", "Category GetCategoryWithItems EmptyFilter")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)
			seedCategoryWithItemsFixture(t, tx, tenantId)

			page := 1
			pagePerContent := 10
			categoryWithItemFromDB, count, err := categoryRepositoryImpl.GetCategoryWithItems(
				tenantId, page-1, pagePerContent, "", 0,
				[]query.QueryFilter{{Column: "", Ascending: true}},
			)
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), "WARN")
			}
			assert.Nil(t, categoryWithItemFromDB)
			assert.Equal(t, 0, count)
		})

		t.Run("NotExistTenantId", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_getcategorywithitems_notexisttenant@example.com", "Category GetCategoryWithItems NotExistTenant")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)
			seedCategoryWithItemsFixture(t, tx, tenantId)

			page := 1
			pagePerContent := 10
			// tenantId+999999 was never seeded, so it must not exist.
			categoryWithItemFromDB, count, err := categoryRepositoryImpl.GetCategoryWithItems(tenantId+999999, page-1, pagePerContent, "", 0, []query.QueryFilter{})
			assert.NoError(t, err)
			assert.Equal(t, 0, count)
			assert.Empty(t, categoryWithItemFromDB)
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("NormalGetAll", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_get_normalgetall@example.com", "Category Get NormalGetAll")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)

			dummyCategories := []*model.Category{
				{TenantId: tenantId, CategoryName: "Category 1 " + uuid.NewString()},
				{TenantId: tenantId, CategoryName: "Category 2 " + uuid.NewString()},
			}
			_, err := categoryRepositoryImpl.Create(tenantId, dummyCategories)
			require.NoError(t, err)

			page := 1
			pagePerContent := 2
			categories, count, err := categoryRepositoryImpl.Get(tenantId, page-1, pagePerContent, "")
			assert.Nil(t, err)
			assert.NotEqual(t, 0, count)
			assert.NotNil(t, categories)
			assert.Greater(t, len(categories), 0)
		})

		t.Run("GetByNameQuerySubstringMatch", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_get_bynamequery_substring@example.com", "Category Get ByNameQuery Substring")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)

			uniqueString := uuid.NewString()
			// The query term sits in the MIDDLE of the name, not at the start —
			// this only matches if the LIKE pattern is %term%, not term%.
			categoryName := "Cold_" + uniqueString + "_Beverages"
			dummyCategory := &model.Category{CategoryName: categoryName, TenantId: tenantId}

			require.NoError(t, tx.Create(dummyCategory).Error)

			categories, count, err := categoryRepositoryImpl.Get(tenantId, 0, 10, uniqueString+"_Beverages")
			assert.NoError(t, err)
			assert.Equal(t, 1, count)
			require.Len(t, categories, 1)
			assert.Equal(t, categoryName, categories[0].CategoryName)
		})

		t.Run("Overflow", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_get_overflow@example.com", "Category Get Overflow")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)

			dummyCategories := []*model.Category{
				{TenantId: tenantId, CategoryName: "Category 1 " + uuid.NewString()},
				{TenantId: tenantId, CategoryName: "Category 2 " + uuid.NewString()},
			}
			_, err := categoryRepositoryImpl.Create(tenantId, dummyCategories)
			require.NoError(t, err)

			// No error because correct page
			page := 1
			pagePerContent := 999
			categories, count, err := categoryRepositoryImpl.Get(tenantId, page-1, pagePerContent, "")
			assert.Greater(t, count, 0)
			assert.Nil(t, err)
			assert.NotNil(t, categories)

			// Page far beyond the seeded data: no error, just empty results.
			page = 100
			pagePerContent = 999
			categories, _, err = categoryRepositoryImpl.Get(tenantId, page-1, pagePerContent, "")
			assert.NoError(t, err)
			assert.Empty(t, categories)
		})

		t.Run("GetByNameQuery", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_get_bynamequery@example.com", "Category Get ByNameQuery")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)

			uniqueString := uuid.NewString()
			namePrefix := "Test_GetByNameQuery_" + uniqueString
			dummyCategories := []*model.Category{
				{CategoryName: namePrefix + " 1", TenantId: tenantId},
				{CategoryName: namePrefix + " 2", TenantId: tenantId},
			}

			// Manually insert without using another repository method
			err := tx.Create(&dummyCategories).Error
			require.NoError(t, err)
			require.NotNil(t, dummyCategories)

			page := 1
			pagePerContent := 10
			categories, count, err := categoryRepositoryImpl.Get(tenantId, page-1, pagePerContent, namePrefix)
			assert.NoError(t, err)
			assert.NotNil(t, categories)
			assert.Equal(t, len(dummyCategories), count)
			require.Len(t, categories, len(dummyCategories))
			for i, category := range categories {
				assert.Equal(t, dummyCategories[i].CategoryName, category.CategoryName)
				assert.Equal(t, dummyCategories[i].TenantId, category.TenantId)
				assert.NotEqual(t, 0, category.Id)
				assert.NotNil(t, category.CreatedAt)
			}
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("CreateOne", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_create_createone@example.com", "Category Create CreateOne")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)

			dummyData := &model.Category{
				TenantId:     tenantId,
				CategoryName: "Test_CategoryRepositoryImpl_Create_CreateOne 1 " + uuid.NewString(),
			}
			createdDummyCategoryFromDB, err := categoryRepositoryImpl.Create(tenantId, []*model.Category{dummyData})
			assert.Nil(t, err)
			assert.NotEqual(t, 0, createdDummyCategoryFromDB[0].Id)
			assert.NotEqual(t, 0, len(createdDummyCategoryFromDB))
			assert.Equal(t, dummyData.CategoryName, createdDummyCategoryFromDB[0].CategoryName)

			// Check using Get method if the data really placed in DB
			var actualCategory model.Category
			err = tx.Where("id", createdDummyCategoryFromDB[0].Id).Take(&actualCategory).Error
			require.Nil(t, err)
			assert.Equal(t, createdDummyCategoryFromDB[0].Id, actualCategory.Id)
			assert.Equal(t, createdDummyCategoryFromDB[0].CategoryName, actualCategory.CategoryName)
		})

		t.Run("CreateMultiple", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_create_createmultiple@example.com", "Category Create CreateMultiple")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)

			uniqueString := uuid.NewString()
			dataDummies := []*model.Category{
				{TenantId: tenantId, CategoryName: "Test_CategoryRepositoryImpl_Update_CreateMultiple 1 " + uniqueString},
				{TenantId: tenantId, CategoryName: "Test_CategoryRepositoryImpl_Update_CreateMultiple 2 " + uniqueString},
				{TenantId: tenantId, CategoryName: "Test_CategoryRepositoryImpl_Update_CreateMultiple 3 " + uniqueString},
				{TenantId: tenantId, CategoryName: "Test_CategoryRepositoryImpl_Update_CreateMultiple 4 " + uniqueString},
				{TenantId: tenantId, CategoryName: "Test_CategoryRepositoryImpl_Update_CreateMultiple 5 " + uniqueString},
			}

			_createdDummyCategoryDB, err := categoryRepositoryImpl.Create(tenantId, dataDummies)
			require.Nil(t, err)
			require.Equal(t, 5, len(_createdDummyCategoryDB))

			for i, createdDummy := range _createdDummyCategoryDB {
				assert.Equal(t, dataDummies[i].CategoryName, createdDummy.CategoryName)
				assert.Equal(t, dataDummies[i].TenantId, createdDummy.TenantId)
			}
		})

		t.Run("CreateWithExistingId", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_create_existingid@example.com", "Category Create ExistingId")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)

			dummyData := &model.Category{
				TenantId:     tenantId,
				CategoryName: "Test_CategoryRepositoryImpl_Update_CreateWithExistingId 1 " + uuid.NewString(),
			}
			createdDummyCategoryFromDB, err := categoryRepositoryImpl.Create(tenantId, []*model.Category{dummyData})
			require.Nil(t, err)
			require.NotEqual(t, 0, createdDummyCategoryFromDB[0].Id)
			require.NotEqual(t, 0, len(createdDummyCategoryFromDB))
			require.Equal(t, dummyData.CategoryName, createdDummyCategoryFromDB[0].CategoryName)

			// Begin test; Assigning Id to Category is illegal insert
			duplicateData := &model.Category{
				Id:           createdDummyCategoryFromDB[0].Id,
				TenantId:     tenantId,
				CategoryName: "Test_CategoryRepositoryImpl_Update_CreateMultiple 1 " + uuid.NewString(),
			}
			duplicateDataFromDB, err := categoryRepositoryImpl.Create(tenantId, []*model.Category{duplicateData})
			assert.Nil(t, duplicateDataFromDB)
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), "23505")
			}
		})

		t.Run("CreateWithExactCategoryName", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_create_exactname@example.com", "Category Create ExactName")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)

			categoryName := "Test_CategoryRepositoryImpl_Update_CreateWithExactCategoryName 1 " + uuid.NewString()
			dummyData := &model.Category{TenantId: tenantId, CategoryName: categoryName}
			createdDummyCategoryFromDB, err := categoryRepositoryImpl.Create(tenantId, []*model.Category{dummyData})
			require.Nil(t, err)
			require.NotEqual(t, 0, createdDummyCategoryFromDB[0].Id)
			require.NotEqual(t, 0, len(createdDummyCategoryFromDB))
			require.Equal(t, dummyData.CategoryName, createdDummyCategoryFromDB[0].CategoryName)

			// Begin test; Exact category name is not allowed
			duplicateData := &model.Category{TenantId: tenantId, CategoryName: categoryName}
			duplicateDataFromDB, err := categoryRepositoryImpl.Create(tenantId, []*model.Category{duplicateData})
			assert.Nil(t, duplicateDataFromDB)
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), "23505")
			}
		})
	})

	t.Run("Register", func(t *testing.T) {
		t.Run("NormalRegister", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_register_normal@example.com", "Category Register Normal")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)
			warehouseRepositoryImpl := NewWarehouseRepositoryImpl(tx)

			dummyItem := &model.Item{
				ItemName:  "Test_CategoryRepositoryImpl_Register_NormalRegister 1 " + uuid.NewString(),
				Stocks:    10,
				TenantId:  tenantId,
				IsActive:  true,
				StockType: model.StockTypeTracked,
			}
			_dummyItemFromDB, err := warehouseRepositoryImpl.CreateItem([]*model.Item{dummyItem})
			require.Nil(t, err)
			require.Equal(t, 1, len(_dummyItemFromDB))
			dummyItemFromDB := _dummyItemFromDB[0]

			dummyCategory := &model.Category{
				CategoryName: "Test_CategoryRepositoryImpl_Register_NormalRegister 1 The Category " + uuid.NewString(),
				TenantId:     tenantId,
			}
			_createdDummyCategoryFromDB, err := categoryRepositoryImpl.Create(dummyCategory.TenantId, []*model.Category{dummyCategory})
			require.Nil(t, err)
			require.Equal(t, 1, len(_createdDummyCategoryFromDB))
			createdDummyCategoryFromDB := _createdDummyCategoryFromDB[0]

			dummyCategoryMtmWarehouse := []*model.CategoryMtmWarehouse{
				{CategoryId: createdDummyCategoryFromDB.Id, ItemId: dummyItemFromDB.ItemId},
			}

			// The test itself
			err = categoryRepositoryImpl.Register(dummyCategoryMtmWarehouse)
			assert.Nil(t, err)
		})

		t.Run("DuplicateRegister", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_register_duplicate@example.com", "Category Register Duplicate")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)
			warehouseRepositoryImpl := NewWarehouseRepositoryImpl(tx)

			dummyItem := &model.Item{
				ItemName:  "Test_CategoryRepositoryImpl_Register_DuplicateRegister 1 " + uuid.NewString(),
				Stocks:    10,
				StockType: model.StockTypeTracked,
				TenantId:  tenantId,
				IsActive:  true,
			}
			_dummyItemFromDB, err := warehouseRepositoryImpl.CreateItem([]*model.Item{dummyItem})
			require.Nil(t, err)
			require.Equal(t, 1, len(_dummyItemFromDB))
			dummyItemFromDB := _dummyItemFromDB[0]

			dummyCategory := &model.Category{
				CategoryName: "Test_CategoryRepositoryImpl_Register_DuplicateRegister 1 The Category " + uuid.NewString(),
				TenantId:     tenantId,
			}
			_createdDummyCategoryFromDB, err := categoryRepositoryImpl.Create(dummyCategory.TenantId, []*model.Category{dummyCategory})
			require.Nil(t, err)
			require.Equal(t, 1, len(_createdDummyCategoryFromDB))
			createdDummyCategoryFromDB := _createdDummyCategoryFromDB[0]

			dummyCategoryMtmWarehouse := []*model.CategoryMtmWarehouse{
				{CategoryId: createdDummyCategoryFromDB.Id, ItemId: dummyItemFromDB.ItemId},
			}

			// Test itself, repeat twice
			err = categoryRepositoryImpl.Register(dummyCategoryMtmWarehouse)
			assert.Nil(t, err)

			err = categoryRepositoryImpl.Register(dummyCategoryMtmWarehouse)
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), "23505")
			}
		})
	})

	t.Run("Unregister", func(t *testing.T) {
		t.Run("NormalUnregister", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_unregister_normal@example.com", "Category Unregister Normal")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)
			warehouseRepositoryImpl := NewWarehouseRepositoryImpl(tx)

			dummyItem := &model.Item{
				ItemName:  "Test_CategoryRepositoryImpl_Register_NormalUnregister 1 " + uuid.NewString(),
				Stocks:    10,
				StockType: model.StockTypeTracked,
				TenantId:  tenantId,
				IsActive:  true,
			}
			_dummyItemFromDB, err := warehouseRepositoryImpl.CreateItem([]*model.Item{dummyItem})
			require.Nil(t, err)
			require.Equal(t, 1, len(_dummyItemFromDB))
			dummyItemFromDB := _dummyItemFromDB[0]

			dummyCategory := &model.Category{
				CategoryName: "Test_CategoryRepositoryImpl_Register_NormalUnregister 1 The Category " + uuid.NewString(),
				TenantId:     tenantId,
			}
			_createdDummyCategoryFromDB, err := categoryRepositoryImpl.Create(dummyCategory.TenantId, []*model.Category{dummyCategory})
			require.Nil(t, err)
			require.Equal(t, 1, len(_createdDummyCategoryFromDB))
			createdDummyCategoryFromDB := _createdDummyCategoryFromDB[0]

			dummyCategoryMtmWarehouse := []*model.CategoryMtmWarehouse{
				{CategoryId: createdDummyCategoryFromDB.Id, ItemId: dummyItemFromDB.ItemId},
			}
			err = categoryRepositoryImpl.Register(dummyCategoryMtmWarehouse)
			assert.Nil(t, err)

			// The test itself
			tobeUnregisteredCategoryMtmWarehouse := &model.CategoryMtmWarehouse{
				CategoryId: createdDummyCategoryFromDB.Id,
				ItemId:     dummyItemFromDB.ItemId,
			}
			err = categoryRepositoryImpl.Unregister(tobeUnregisteredCategoryMtmWarehouse)
			assert.Nil(t, err)
		})

		t.Run("UnregisterThatUnregistered", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)

			tobeUnregisteredCategoryMtmWarehouse1 := &model.CategoryMtmWarehouse{CategoryId: 0, ItemId: 1}
			tobeUnregisteredCategoryMtmWarehouse2 := &model.CategoryMtmWarehouse{CategoryId: 1, ItemId: 0}

			err := categoryRepositoryImpl.Unregister(tobeUnregisteredCategoryMtmWarehouse1)
			require.NotNil(t, err, "If this fail, immediately check the the test data; there is a possibility data deleted !")
			require.Contains(t, err.Error(), "[WARN]")

			err = categoryRepositoryImpl.Unregister(tobeUnregisteredCategoryMtmWarehouse2)
			require.NotNil(t, err, "If this fail, immediately check the the test data; there is a possibility data deleted !")
			require.Contains(t, err.Error(), "[WARN]")
		})
	})

	t.Run("EditItemCategory", func(t *testing.T) {
		t.Run("NormalEditItemCategory", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_edititemcategory_normal@example.com", "Category EditItemCategory Normal")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)
			warehouseRepositoryImpl := NewWarehouseRepositoryImpl(tx)

			dummyItem := &model.Item{
				ItemName:  "Test_CategoryRepositoryImpl_EditItemCategory_NormalEditItemCategory 1 " + uuid.NewString(),
				Stocks:    10,
				StockType: model.StockTypeTracked,
				TenantId:  tenantId,
				IsActive:  true,
			}
			_dummyItemFromDB, err := warehouseRepositoryImpl.CreateItem([]*model.Item{dummyItem})
			require.Nil(t, err)
			require.Equal(t, 1, len(_dummyItemFromDB))
			dummyItemFromDB := _dummyItemFromDB[0]

			dummyCategories := []*model.Category{
				{CategoryName: "Test_CategoryRepositoryImpl_EditItemCategory_NormalEditItemCategory 1 The Category " + uuid.NewString(), TenantId: tenantId},
				{CategoryName: "Test_CategoryRepositoryImpl_EditItemCategory_NormalEditItemCategory 2 The Category " + uuid.NewString(), TenantId: tenantId},
			}
			_createdDummyCategoryFromDB, err := categoryRepositoryImpl.Create(tenantId, dummyCategories)
			require.Nil(t, err)
			require.Len(t, _createdDummyCategoryFromDB, len(dummyCategories))

			createdDummyCategoryFromDB1 := _createdDummyCategoryFromDB[0]
			createdDummyCategoryFromDB2 := _createdDummyCategoryFromDB[1]

			dummyCategoryMtmWarehouse := []*model.CategoryMtmWarehouse{
				{CategoryId: createdDummyCategoryFromDB1.Id, ItemId: dummyItemFromDB.ItemId},
			}
			err = categoryRepositoryImpl.Register(dummyCategoryMtmWarehouse)
			require.Nil(t, err)

			// The test itself
			tobeUpdateItemCategory := &model.CategoryMtmWarehouse{
				CategoryId: createdDummyCategoryFromDB2.Id, // Edited
				ItemId:     dummyItemFromDB.ItemId,
			}
			err = categoryRepositoryImpl.EditItemCategory(tenantId, tobeUpdateItemCategory)
			assert.NoError(t, err)
		})

		t.Run("NonExistenceItemAtWarehouse", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_edititemcategory_noitem@example.com", "Category EditItemCategory NoItem")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)

			// Expected to get this error message
			// [ERROR] Fatal error, current item from store never exist at warehouse
			tobeUpdateItemCategory := &model.CategoryMtmWarehouse{
				CategoryId: 1,        // Technically valid
				ItemId:     99999999, // Not available ItemId
			}
			err := categoryRepositoryImpl.EditItemCategory(tenantId, tobeUpdateItemCategory)
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), "[ERROR]")
			}
		})

		t.Run("InvalidRequest", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_edititemcategory_invalid@example.com", "Category EditItemCategory Invalid")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)

			// User will act as unregister by sending categoryId: 0
			tobeUpdateItemCategory := &model.CategoryMtmWarehouse{
				CategoryId: 0, // Invalid input request
				ItemId:     1, // Technically valid
			}
			err := categoryRepositoryImpl.EditItemCategory(tenantId, tobeUpdateItemCategory)
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), "[ERROR]")
			}
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("NormalUpdate", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_update_normal@example.com", "Category Update Normal")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)

			dummyData := &model.Category{
				TenantId:     tenantId,
				CategoryName: "Test_CategoryRepositoryImpl_Update_NormalUpdate 1 " + uuid.NewString(),
			}
			_createdDummyCategoryDB, err := categoryRepositoryImpl.Create(tenantId, []*model.Category{dummyData})
			require.Nil(t, err)
			require.Equal(t, 1, len(_createdDummyCategoryDB))

			// Begin updating; only CategoryName could be updated
			createdDummyCategoryDB := _createdDummyCategoryDB[0]
			createdDummyCategoryDB.CategoryName = "Test_CategoryRepositoryImpl_Update_NormalUpdate 1 (UPDATED) " + uuid.NewString()

			editedDummyCategoryDB, err := categoryRepositoryImpl.Update(createdDummyCategoryDB.TenantId, createdDummyCategoryDB.Id, createdDummyCategoryDB.CategoryName)
			assert.Nil(t, err)
			assert.NotNil(t, editedDummyCategoryDB)
			assert.Equal(t, createdDummyCategoryDB.Id, editedDummyCategoryDB.Id)
			assert.Equal(t, createdDummyCategoryDB.TenantId, editedDummyCategoryDB.TenantId)
			assert.Equal(t, createdDummyCategoryDB.CategoryName, editedDummyCategoryDB.CategoryName)
			assert.Equal(t, createdDummyCategoryDB.CreatedAt.UTC().Day(), editedDummyCategoryDB.CreatedAt.UTC().Day())
		})

		t.Run("UpdateThatCategoryNotExist", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_update_notexist@example.com", "Category Update NotExist")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)

			editedDummyCategoryDB, err := categoryRepositoryImpl.Update(tenantId, 0, "Will not happen")
			assert.NotNil(t, err)
			assert.Nil(t, editedDummyCategoryDB)
			if assert.Error(t, err) {
				assert.Equal(t, "record not found", err.Error())
			}
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("NormalDelete", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_delete_normal@example.com", "Category Delete Normal")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)

			dummyData := &model.Category{
				TenantId:     tenantId,
				CategoryName: "Test_CategoryRepositoryImpl_Delete_NormalDelete 1 " + uuid.NewString(),
			}
			_createdDummyCategoryDB, err := categoryRepositoryImpl.Create(tenantId, []*model.Category{dummyData})
			require.Nil(t, err)
			require.Equal(t, 1, len(_createdDummyCategoryDB))

			createdDummyCategoryDB := _createdDummyCategoryDB[0]

			err = categoryRepositoryImpl.Delete(createdDummyCategoryDB)
			assert.Nil(t, err)
		})

		t.Run("DeleteThatCategoryNotExist", func(t *testing.T) {
			tx := gormClient.Begin()
			defer tx.Rollback()

			tenantId, _ := seedOrderItemTestDependencies(t, tx, "category_delete_notexist@example.com", "Category Delete NotExist")
			categoryRepositoryImpl := NewCategoryRepositoryImpl(tx)

			dummyData := &model.Category{
				Id:           0,
				TenantId:     tenantId,
				CategoryName: "Test_CategoryRepositoryImpl_Delete_DeleteThatCategoryNotExist 1",
			}
			err := categoryRepositoryImpl.Delete(dummyData)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "[WARN]")
		})
	})
}
