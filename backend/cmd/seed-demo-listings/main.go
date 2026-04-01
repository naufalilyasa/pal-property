package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	requestdto "github.com/naufalilyasa/pal-property-backend/internal/dto/request"
	"github.com/naufalilyasa/pal-property-backend/internal/repository/postgres"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	"github.com/naufalilyasa/pal-property-backend/pkg/authz"
	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"github.com/naufalilyasa/pal-property-backend/pkg/logger"
	"github.com/naufalilyasa/pal-property-backend/pkg/searchindex"
	"go.uber.org/zap"
	pgDriver "gorm.io/driver/postgres"
	gormPkg "gorm.io/gorm"
)

type regionPath struct {
	ProvinceCode string
	CityCode     string
	DistrictCode string
	VillageCode  string
}

type demoListingSeed struct {
	Title             string
	Slug              string
	Description       string
	CategorySlug      string
	TransactionType   string
	Price             int64
	SpecialOffers     []string
	Region            regionPath
	AddressDetail     string
	Latitude          float64
	Longitude         float64
	Bedrooms          int
	Bathrooms         int
	FloorCount        int
	CarportCapacity   int
	LandAreaSqm       int
	BuildingAreaSqm   int
	CertificateType   string
	Condition         string
	Furnishing        string
	ElectricalPowerVA int
	FacingDirection   string
	YearBuilt         int
	Facilities        []string
	Status            string
	Images            []demoMedia
	Video             demoVideo
}

type demoMedia struct {
	URL              string
	OriginalFilename string
	Width            int
	Height           int
	Bytes            int64
	Format           string
}

type demoVideo struct {
	URL              string
	OriginalFilename string
	Width            int
	Height           int
	Bytes            int64
	Format           string
	DurationSeconds  int
}

type appUser struct {
	ID    uuid.UUID
	Email string
	Role  string
}

type appCategory struct {
	ID   uuid.UUID
	Slug string
}

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	logger.InitLogger()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.Env.DBHost, config.Env.DBUser, config.Env.DBPassword,
		config.Env.DBName, config.Env.DBPort, config.Env.DBSSLMode,
	)

	db, err := gormPkg.Open(pgDriver.Open(dsn), &gormPkg.Config{})
	if err != nil {
		logger.Log.Fatal("Failed to connect to database", zap.Error(err))
	}

	ctx := context.Background()
	if err := postgres.EnsureIndonesiaRegionsSeeded(ctx, db, config.Env.WilayahDataPath); err != nil {
		logger.Log.Fatal("Failed to seed wilayah data", zap.Error(err))
	}

	owner, err := findSeedOwner(ctx, db)
	if err != nil {
		logger.Log.Fatal("Failed to find seed owner", zap.Error(err))
	}

	listingRepo := postgres.NewListingRepository(db)
	indexJobRepo := postgres.NewSearchIndexJobRepository(db)
	indexTxManager := postgres.NewSearchIndexTransactionManager(db)
	listingAuthzService := service.NewAuthzService(nil)
	regionService := service.NewRegionService(postgres.NewRegionRepository(db))
	listingService := service.NewListingServiceWithAuthzJobsAndTransactions(listingRepo, listingAuthzService, indexJobRepo, indexTxManager)
	listingService = service.WithRegionLookupService(listingService, regionService)
	principal := authz.Principal{UserID: owner.ID, Role: owner.Role}

	categoryIDs, err := loadCategoryIDs(ctx, db)
	if err != nil {
		logger.Log.Fatal("Failed to load categories", zap.Error(err))
	}

	seeds := buildDemoListings()
	if err := purgeExistingDemoListings(ctx, db, seeds); err != nil {
		logger.Log.Fatal("Failed to purge existing demo listings", zap.Error(err))
	}
	createdCount := 0

	for _, seed := range seeds {
		listingReq, err := buildCreateListingRequest(ctx, db, seed, categoryIDs)
		if err != nil {
			logger.Log.Fatal("Failed to build listing request", zap.String("slug", seed.Slug), zap.Error(err))
		}

		created, err := listingService.Create(ctx, principal, listingReq)
		if err != nil {
			logger.Log.Fatal("Failed to create demo listing", zap.String("slug", seed.Slug), zap.Error(err))
		}

		if err := seedListingMedia(ctx, db, created.ID, seed); err != nil {
			logger.Log.Fatal("Failed to seed listing media", zap.String("slug", seed.Slug), zap.Error(err))
		}

		createdCount++
	}

	searchClient, err := searchindex.NewClient(config.Env.ElasticAddress, config.Env.ElasticUsername, config.Env.ElasticPassword, nil)
	if err != nil {
		logger.Log.Fatal("Failed to initialize search client", zap.Error(err))
	}

	if err := service.RebuildListingIndex(ctx, listingRepo, searchClient, config.Env.ElasticListingsIndex, 200); err != nil {
		logger.Log.Fatal("Failed to rebuild listing index after demo seed", zap.Error(err))
	}

	logger.Log.Info("Demo listing seed complete", zap.Int("created", createdCount))
}

func purgeExistingDemoListings(ctx context.Context, db *gormPkg.DB, seeds []demoListingSeed) error {
	titles := make([]string, 0, len(seeds))
	for _, seed := range seeds {
		titles = append(titles, seed.Title)
	}

	var listingIDs []uuid.UUID
	if err := db.WithContext(ctx).
		Table("listings").
		Where("title IN ?", titles).
		Pluck("id", &listingIDs).Error; err != nil {
		return err
	}
	if len(listingIDs) == 0 {
		return nil
	}

	if err := db.WithContext(ctx).Where("listing_id IN ?", listingIDs).Delete(&entity.ListingVideo{}).Error; err != nil {
		return err
	}
	if err := db.WithContext(ctx).Where("listing_id IN ?", listingIDs).Delete(&entity.ListingImage{}).Error; err != nil {
		return err
	}
	if err := db.WithContext(ctx).Unscoped().Where("id IN ?", listingIDs).Delete(&entity.Listing{}).Error; err != nil {
		return err
	}

	return nil
}

func findSeedOwner(ctx context.Context, db *gormPkg.DB) (*appUser, error) {
	var user appUser
	if err := db.WithContext(ctx).
		Table("users").
		Select("id, email, role").
		Order("created_at ASC").
		Take(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func loadCategoryIDs(ctx context.Context, db *gormPkg.DB) (map[string]uuid.UUID, error) {
	var rows []appCategory
	if err := db.WithContext(ctx).Table("categories").Select("id, slug").Find(&rows).Error; err != nil {
		return nil, err
	}

	categoryIDs := make(map[string]uuid.UUID, len(rows))
	for _, row := range rows {
		categoryIDs[row.Slug] = row.ID
	}
	return categoryIDs, nil
}

func buildCreateListingRequest(ctx context.Context, db *gormPkg.DB, seed demoListingSeed, categoryIDs map[string]uuid.UUID) (*requestdto.CreateListingRequest, error) {
	categoryID, ok := categoryIDs[seed.CategorySlug]
	if !ok {
		return nil, fmt.Errorf("category slug %s not found", seed.CategorySlug)
	}

	provinceCode, provinceName, err := findRegionByCode(ctx, db, 1, seed.Region.ProvinceCode, nil)
	if err != nil {
		return nil, err
	}
	cityCode, cityName, err := findRegionByCode(ctx, db, 2, seed.Region.CityCode, &provinceCode)
	if err != nil {
		return nil, err
	}
	districtCode, districtName, err := findRegionByCode(ctx, db, 3, seed.Region.DistrictCode, &cityCode)
	if err != nil {
		return nil, err
	}
	villageCode, villageName, err := findRegionByCode(ctx, db, 4, seed.Region.VillageCode, &districtCode)
	if err != nil {
		return nil, err
	}

	categoryIDPtr := categoryID
	price := seed.Price
	latitude := seed.Latitude
	longitude := seed.Longitude
	bedrooms := seed.Bedrooms
	bathrooms := seed.Bathrooms
	floorCount := seed.FloorCount
	carportCapacity := seed.CarportCapacity
	landArea := seed.LandAreaSqm
	buildingArea := seed.BuildingAreaSqm
	electricalPower := seed.ElectricalPowerVA
	yearBuilt := seed.YearBuilt
	transactionType := seed.TransactionType
	currency := "IDR"
	certificateType := seed.CertificateType
	condition := seed.Condition
	furnishing := seed.Furnishing
	facingDirection := seed.FacingDirection
	status := seed.Status
	isNegotiable := true
	title := seed.Title
	description := seed.Description
	addressDetail := seed.AddressDetail
	provinceCodePtr := provinceCode
	provinceNamePtr := provinceName
	cityCodePtr := cityCode
	cityNamePtr := cityName
	districtCodePtr := districtCode
	districtNamePtr := districtName
	villageCodePtr := villageCode
	villageNamePtr := villageName

	return &requestdto.CreateListingRequest{
		CategoryID:           &categoryIDPtr,
		Title:                title,
		Description:          &description,
		TransactionType:      transactionType,
		Price:                price,
		Currency:             &currency,
		IsNegotiable:         &isNegotiable,
		SpecialOffers:        seed.SpecialOffers,
		LocationProvince:     &provinceNamePtr,
		LocationProvinceCode: &provinceCodePtr,
		LocationCity:         &cityNamePtr,
		LocationCityCode:     &cityCodePtr,
		LocationDistrict:     &districtNamePtr,
		LocationDistrictCode: &districtCodePtr,
		LocationVillage:      &villageNamePtr,
		LocationVillageCode:  &villageCodePtr,
		AddressDetail:        &addressDetail,
		Latitude:             &latitude,
		Longitude:            &longitude,
		BedroomCount:         &bedrooms,
		BathroomCount:        &bathrooms,
		FloorCount:           &floorCount,
		CarportCapacity:      &carportCapacity,
		LandAreaSqm:          &landArea,
		BuildingAreaSqm:      &buildingArea,
		CertificateType:      &certificateType,
		Condition:            &condition,
		Furnishing:           &furnishing,
		ElectricalPowerVA:    &electricalPower,
		FacingDirection:      &facingDirection,
		YearBuilt:            &yearBuilt,
		Facilities:           seed.Facilities,
		Status:               status,
	}, nil
}

func findRegionByCode(ctx context.Context, db *gormPkg.DB, level int, code string, parentCode *string) (string, string, error) {
	query := db.WithContext(ctx).
		Table("indonesia_regions").
		Select("code, name").
		Where("level = ?", level).
		Where("code = ?", code)
	if parentCode != nil {
		query = query.Where("parent_code = ?", *parentCode)
	}

	var region struct {
		Code string
		Name string
	}
	if err := query.Take(&region).Error; err != nil {
		return "", "", fmt.Errorf("find region code %s: %w", code, err)
	}
	return region.Code, region.Name, nil
}

func seedListingMedia(ctx context.Context, db *gormPkg.DB, listingID uuid.UUID, seed demoListingSeed) error {
	for index, image := range seed.Images {
		imageRow := entity.ListingImage{
			ListingID:        listingID,
			URL:              image.URL,
			AssetID:          stringPointer(fmt.Sprintf("demo-image-asset-%s-%02d", seed.Slug, index+1)),
			PublicID:         stringPointer(fmt.Sprintf("demo/%s/image-%02d", seed.Slug, index+1)),
			Version:          int64Pointer(1),
			ResourceType:     stringPointer("image"),
			Type:             stringPointer("upload"),
			Format:           stringPointer(image.Format),
			Bytes:            int64Pointer(image.Bytes),
			Width:            intPointer(image.Width),
			Height:           intPointer(image.Height),
			OriginalFilename: stringPointer(image.OriginalFilename),
			IsPrimary:        index == 0,
			SortOrder:        index,
		}
		if err := db.WithContext(ctx).Create(&imageRow).Error; err != nil {
			return fmt.Errorf("create image %d for %s: %w", index+1, seed.Slug, err)
		}
	}

	videoRow := map[string]any{
		"listing_id":        listingID,
		"url":               seed.Video.URL,
		"asset_id":          fmt.Sprintf("demo-video-asset-%s", seed.Slug),
		"public_id":         fmt.Sprintf("demo/%s/video", seed.Slug),
		"version":           int64(1),
		"resource_type":     "video",
		"delivery_type":     "upload",
		"format":            seed.Video.Format,
		"bytes":             seed.Video.Bytes,
		"width":             seed.Video.Width,
		"height":            seed.Video.Height,
		"original_filename": seed.Video.OriginalFilename,
	}
	if err := db.WithContext(ctx).Table("listing_videos").Create(videoRow).Error; err != nil {
		return fmt.Errorf("create video for %s: %w", seed.Slug, err)
	}

	return nil
}

func intPointer(value int) *int          { return &value }
func int64Pointer(value int64) *int64    { return &value }
func stringPointer(value string) *string { return &value }

func buildDemoListings() []demoListingSeed {
	sharedVideo := demoVideo{
		URL:              "https://assets.mixkit.co/videos/preview/mixkit-modern-house-exterior-view-1171-large.mp4",
		OriginalFilename: "property-tour.mp4",
		Width:            1920,
		Height:           1080,
		Bytes:            12800000,
		Format:           "mp4",
		DurationSeconds:  46,
	}

	return []demoListingSeed{
		{
			Title:             "Rumah Modern Senopati dengan Kolam Renang Privat",
			Slug:              "demo-rumah-modern-senopati-kolam-renang",
			Description:       "Rumah modern tiga lantai di jantung Senopati dengan pencahayaan alami, kolam renang privat, dan ruang keluarga luas. Cocok untuk keluarga mapan yang mengutamakan akses cepat ke SCBD dan area kuliner premium.",
			CategorySlug:      "rumah",
			TransactionType:   "sale",
			Price:             28500000000,
			SpecialOffers:     []string{"Promo", "Turun_Harga"},
			Region:            regionPath{ProvinceCode: "31", CityCode: "31.74", DistrictCode: "31.74.07", VillageCode: "31.74.07.1006"},
			AddressDetail:     "Jl. Senopati Raya No. 18, dekat SCBD dan area premium dining.",
			Latitude:          -6.2279,
			Longitude:         106.8098,
			Bedrooms:          5,
			Bathrooms:         5,
			FloorCount:        3,
			CarportCapacity:   3,
			LandAreaSqm:       420,
			BuildingAreaSqm:   560,
			CertificateType:   "SHM",
			Condition:         "new",
			Furnishing:        "fully",
			ElectricalPowerVA: 22000,
			FacingDirection:   "north",
			YearBuilt:         2023,
			Facilities:        []string{"AC", "CCTV", "Wifi", "Water_Heater", "Carport", "Garden", "Pool", "Security"},
			Status:            "active",
			Images:            buildDemoImages("senopati-modern-house", 0),
			Video:             sharedVideo,
		},
		{
			Title:             "Apartemen Premium Setiabudi Sky Residence Full Furnished",
			Slug:              "demo-apartemen-premium-setiabudi-sky-residence",
			Description:       "Unit apartemen high-rise full furnished dengan view city skyline, private lift access, dan fasilitas premium. Ideal untuk eksekutif yang membutuhkan hunian representatif dekat Kuningan dan Sudirman.",
			CategorySlug:      "apartemen",
			TransactionType:   "sale",
			Price:             7200000000,
			SpecialOffers:     []string{"Promo", "DP_0"},
			Region:            regionPath{ProvinceCode: "31", CityCode: "31.74", DistrictCode: "31.74.02", VillageCode: "31.74.02.1005"},
			AddressDetail:     "Tower A lantai tinggi, akses cepat ke Kuningan dan Sudirman Central Business District.",
			Latitude:          -6.2207,
			Longitude:         106.8296,
			Bedrooms:          3,
			Bathrooms:         2,
			FloorCount:        1,
			CarportCapacity:   1,
			LandAreaSqm:       0,
			BuildingAreaSqm:   178,
			CertificateType:   "Strata",
			Condition:         "second",
			Furnishing:        "fully",
			ElectricalPowerVA: 7700,
			FacingDirection:   "east",
			YearBuilt:         2019,
			Facilities:        []string{"AC", "CCTV", "Wifi", "Water_Heater", "Gym", "Pool", "Security"},
			Status:            "active",
			Images:            buildDemoImages("setiabudi-sky-apartment", 1),
			Video:             sharedVideo,
		},
		{
			Title:             "Rumah Bogor Tengah Dekat Kebun Raya dan Sekolah Favorit",
			Slug:              "demo-rumah-bogor-tengah-dekat-kebun-raya",
			Description:       "Rumah keluarga di Bogor Tengah dengan halaman depan rapi, akses cepat ke Kebun Raya, dan lingkungan yang nyaman untuk keluarga mapan. Cocok untuk pembeli yang mencari rumah siap huni di pusat kota Bogor.",
			CategorySlug:      "rumah",
			TransactionType:   "sale",
			Price:             5450000000,
			SpecialOffers:     []string{"Promo", "Turun_Harga"},
			Region:            regionPath{ProvinceCode: "32", CityCode: "32.71", DistrictCode: "32.71.03", VillageCode: "32.71.03.1003"},
			AddressDetail:     "Area strategis Bogor Tengah dengan akses singkat ke sekolah, rumah sakit, dan pusat kuliner keluarga.",
			Latitude:          -6.5942,
			Longitude:         106.7975,
			Bedrooms:          4,
			Bathrooms:         4,
			FloorCount:        2,
			CarportCapacity:   2,
			LandAreaSqm:       260,
			BuildingAreaSqm:   320,
			CertificateType:   "SHM",
			Condition:         "second",
			Furnishing:        "semi",
			ElectricalPowerVA: 10600,
			FacingDirection:   "south",
			YearBuilt:         2017,
			Facilities:        []string{"AC", "CCTV", "Wifi", "Water_Heater", "Carport", "Garden", "Security"},
			Status:            "active",
			Images:            buildDemoImages("bogor-family-house", 2),
			Video:             sharedVideo,
		},
		{
			Title:             "Townhouse Cibubur Dekat Tol dan LRT",
			Slug:              "demo-townhouse-cibubur-dekat-tol-dan-lrt",
			Description:       "Townhouse modern minimalis di kawasan Cibubur dengan akses cepat ke tol, LRT, dan pusat belanja. Cocok untuk keluarga muda yang mencari rumah praktis dengan lokasi strategis.",
			CategorySlug:      "townhouse",
			TransactionType:   "sale",
			Price:             3950000000,
			SpecialOffers:     []string{"DP_0", "Promo"},
			Region:            regionPath{ProvinceCode: "32", CityCode: "32.76", DistrictCode: "32.76.02", VillageCode: "32.76.02.1007"},
			AddressDetail:     "Cluster premium dekat Gerbang Tol Cimanggis dan stasiun LRT Harjamukti.",
			Latitude:          -6.3652,
			Longitude:         106.9015,
			Bedrooms:          4,
			Bathrooms:         3,
			FloorCount:        2,
			CarportCapacity:   2,
			LandAreaSqm:       144,
			BuildingAreaSqm:   210,
			CertificateType:   "SHM",
			Condition:         "new",
			Furnishing:        "semi",
			ElectricalPowerVA: 6600,
			FacingDirection:   "west",
			YearBuilt:         2024,
			Facilities:        []string{"AC", "CCTV", "Wifi", "Carport", "Playground", "Security"},
			Status:            "active",
			Images:            buildDemoImages("cibubur-townhouse", 3),
			Video:             sharedVideo,
		},
		{
			Title:             "Ruko Ciputat Timur Strategis untuk Klinik dan Kantor",
			Slug:              "demo-ruko-ciputat-timur-strategis-klinik-kantor",
			Description:       "Ruko tiga lantai di koridor bisnis Ciputat Timur dengan frontage lebar, ideal untuk klinik, kantor cabang, atau showroom. Lokasinya dekat akses utama BSD, Bintaro, dan Lebak Bulus.",
			CategorySlug:      "ruko",
			TransactionType:   "rent",
			Price:             385000000,
			SpecialOffers:     []string{"Promo"},
			Region:            regionPath{ProvinceCode: "36", CityCode: "36.74", DistrictCode: "36.74.05", VillageCode: "36.74.05.1001"},
			AddressDetail:     "Koridor bisnis aktif Ciputat Timur, dekat akses ke Bintaro, Lebak Bulus, dan area pendidikan.",
			Latitude:          -6.3088,
			Longitude:         106.7652,
			Bedrooms:          0,
			Bathrooms:         3,
			FloorCount:        3,
			CarportCapacity:   4,
			LandAreaSqm:       220,
			BuildingAreaSqm:   480,
			CertificateType:   "HGB",
			Condition:         "second",
			Furnishing:        "unfurnished",
			ElectricalPowerVA: 23000,
			FacingDirection:   "north",
			YearBuilt:         2018,
			Facilities:        []string{"AC", "Wifi", "Carport", "Security"},
			Status:            "active",
			Images:            buildDemoImages("ciputat-commercial-shophouse", 4),
			Video:             sharedVideo,
		},
	}
}

func buildDemoImages(seed string, primaryVariant int) []demoMedia {
	primarySources := []string{
		"https://images.unsplash.com/photo-1568605114967-8130f3a36994?auto=format&fit=crop&w=1600&q=80",
		"https://images.unsplash.com/photo-1600585154526-990dced4db0d?auto=format&fit=crop&w=1600&q=80",
		"https://images.unsplash.com/photo-1512917774080-9991f1c4c750?auto=format&fit=crop&w=1600&q=80",
		"https://images.unsplash.com/photo-1505693416388-ac5ce068fe85?auto=format&fit=crop&w=1600&q=80",
		"https://images.unsplash.com/photo-1484154218962-a197022b5858?auto=format&fit=crop&w=1600&q=80",
	}
	primaryURL := fmt.Sprintf("%s&seed=%s-primary", primarySources[primaryVariant%len(primarySources)], seed)

	return []demoMedia{
		{URL: primaryURL, OriginalFilename: seed + "-01.jpg", Width: 1600, Height: 1067, Bytes: 420000, Format: "jpg"},
		{URL: fmt.Sprintf("https://images.unsplash.com/photo-1600585154526-990dced4db0d?auto=format&fit=crop&w=1600&q=80&seed=%s-02", seed), OriginalFilename: seed + "-02.jpg", Width: 1600, Height: 1067, Bytes: 438000, Format: "jpg"},
		{URL: fmt.Sprintf("https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?auto=format&fit=crop&w=1600&q=80&seed=%s-03", seed), OriginalFilename: seed + "-03.jpg", Width: 1600, Height: 1067, Bytes: 441000, Format: "jpg"},
		{URL: fmt.Sprintf("https://images.unsplash.com/photo-1600047509807-ba8f99d2cdde?auto=format&fit=crop&w=1600&q=80&seed=%s-04", seed), OriginalFilename: seed + "-04.jpg", Width: 1600, Height: 1067, Bytes: 417000, Format: "jpg"},
		{URL: fmt.Sprintf("https://images.unsplash.com/photo-1600573472592-401b489a3cdc?auto=format&fit=crop&w=1600&q=80&seed=%s-05", seed), OriginalFilename: seed + "-05.jpg", Width: 1600, Height: 1067, Bytes: 430000, Format: "jpg"},
		{URL: fmt.Sprintf("https://images.unsplash.com/photo-1600566753151-384129cf4e3e?auto=format&fit=crop&w=1600&q=80&seed=%s-06", seed), OriginalFilename: seed + "-06.jpg", Width: 1600, Height: 1067, Bytes: 426000, Format: "jpg"},
		{URL: fmt.Sprintf("https://images.unsplash.com/photo-1600607687644-c7171b42498f?auto=format&fit=crop&w=1600&q=80&seed=%s-07", seed), OriginalFilename: seed + "-07.jpg", Width: 1600, Height: 1067, Bytes: 432000, Format: "jpg"},
		{URL: fmt.Sprintf("https://images.unsplash.com/photo-1512917774080-9991f1c4c750?auto=format&fit=crop&w=1600&q=80&seed=%s-08", seed), OriginalFilename: seed + "-08.jpg", Width: 1600, Height: 1067, Bytes: 415000, Format: "jpg"},
		{URL: fmt.Sprintf("https://images.unsplash.com/photo-1505693416388-ac5ce068fe85?auto=format&fit=crop&w=1600&q=80&seed=%s-09", seed), OriginalFilename: seed + "-09.jpg", Width: 1600, Height: 1067, Bytes: 423000, Format: "jpg"},
		{URL: fmt.Sprintf("https://images.unsplash.com/photo-1484154218962-a197022b5858?auto=format&fit=crop&w=1600&q=80&seed=%s-10", seed), OriginalFilename: seed + "-10.jpg", Width: 1600, Height: 1067, Bytes: 419000, Format: "jpg"},
	}
}
