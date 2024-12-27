package hiDm

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	dameng "github.com/sunzip/gorm-dameng"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormlog "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/hints"
)

const DM = "DM"

var (
	data *Data
	err  error
)

type Data struct {
	db     *gorm.DB
	dbInfo *Options
}

// Options options
type Options struct {
	// Address address
	Address string
	// UserName username
	UserName string
	// Password password
	Password string
	// DBName dbname
	DBName  string
	Driver  string
	Logger  gormlog.Interface
	Charset string
}

func init() {
	db, err := New("SYSDBA", "SYSDBA", "192.168.20.65", 5236, 0, "test_cloud_core", nil, nil, &gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		panic(err)
	}

	data = &Data{}

	data.db = db
}

func initx() {
	opt := &Options{
		Address:  "192.168.18.140:5236",
		UserName: "SYSDBA",
		Password: "SYSDBA",
		DBName:   "uav",
		Driver:   DM,
	}

	opt.Address = "192.168.20.65:5236"
	opt.DBName = "test_cloud_core"

	// loggerOpt := hiZap.Options{
	// 	Level: zapcore.DebugLevel,
	// 	Skip:  2,
	// }
	// zap 开启stack trace logger, 用于service层输出
	// stackLogger := hiZap.New(&loggerOpt)
	// logger := log.With(hiZap.New(&loggerOpt), hiCommon.TraceID, tracing.TraceID())

	// opt.Logger = NewLog(logger, &gormlog.Config{
	// 	SlowThreshold:             10 * time.Millisecond, // Slow SQL threshold
	// 	LogLevel:                  gormlog.Info,          // Log level
	// 	IgnoreRecordNotFoundError: true,                  // Ignore ErrRecordNotFound error for logger
	// 	Colorful:                  false,                 // Disable color
	// })

	options := map[string]string{
		"schema":         "test_cloud_core",
		"appName":        "GORM 连接达梦数据库示例",
		"connectTimeout": "30000",
	}

	// dm://user:password@host:port?schema=SYSDBA[&...]
	dsn := dameng.BuildUrl("SYSDBA", "SYSDBA", "192.168.20.65", 5236, options)
	cfg := &gorm.Config{Logger: opt.Logger}
	if opt.Logger == nil {
		cfg.Logger = gormlog.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gormlog.Config{
				SlowThreshold:             time.Second,  // Slow SQL threshold
				LogLevel:                  gormlog.Info, // Log level
				IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
				Colorful:                  false,        // Disable color
			},
		)
	}
	cfg.NamingStrategy = schema.NamingStrategy{SingularTable: true}
	db, err := gorm.Open(dameng.Open(dsn), cfg)
	if err != nil {
		panic(err)
	}

	data = &Data{dbInfo: opt}

	data.db = db
}

func (d *Data) DB(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}

type TestClauseConflict struct {
	ID        int64     `json:"id" gorm:"id"`
	UserId    int64     `json:"userId" gorm:"user_id"`
	UserName  string    `json:"userName" gorm:"user_name"`
	DeletedAt time.Time `json:"deletedAt" gorm:"deleted_at"`
	DeletedBy int64     `json:"deletedBy" gorm:"deleted_by"`
	CreatedAt time.Time `json:"createdAt" gorm:"created_at"`
	CreatedBy int64     `json:"createdBy" gorm:"created_by"`
}

// TableExist
//
//	@param tableName
//	@return bool true=存在
//	@return error
func TableExist(tableName string) (bool, error) {
	type Count struct {
		C int
	}
	c := &Count{}
	err := data.DB(context.Background()).Raw(`SELECT COUNT(*) c FROM USER_TABLES WHERE TABLE_NAME = ?;`, tableName).Scan(c).Error
	if err != nil {
		fmt.Println("Create table  :", err)
	}
	fmt.Println("exist:", c.C > 0)
	return c.C > 0, err
}

var tableScript = map[string]string{"test_clause_conflict": `create table test_clause_conflict(id BIGINT NOT NULL AUTO_INCREMENT,user_id BIGINT,user_name varchar(100),deleted_at TIMESTAMP,deleted_by BIGINT,created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,created_by BIGINT NULL,	
	CONSTRAINT "test_clause_conflict_PK" PRIMARY KEY ("id"));`}

func createTestTable(tableName string) error {
	err := data.DB(context.Background()).Exec(tableScript[tableName]).Error
	if err != nil {
		fmt.Println("Create table  :", err)
	}
	return err
}

func existOrCreateTable(t *testing.T, tableName string) {
	exist, e := TableExist(tableName)
	if e != nil {
		t.Fail()
	}
	if !exist {
		e := createTestTable(tableName)
		if e != nil {
			t.Fail()
		}
	}
}

// 其他字段
func TestSaveClauseConflictGitHub2(t *testing.T) {
	{
		existOrCreateTable(t, "test_clause_conflict")
	}

	// test 和Columns 都带 id，是会更新的
	test := &TestClauseConflict{
		UserId: 100,
		// UserName:  "updater" + time.Now().String(),
		UserName:  "5",
		CreatedBy: 111,
		CreatedAt: time.Now(),
	}

	if err := data.db.Model(&TestClauseConflict{}).Debug().Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "user_name"}},

		DoUpdates: clause.AssignmentColumns([]string{
			"created_by", "created_at",
		}),
	}).Create(&test).Error; err != nil {
		t.Fail()
	}

	fmt.Println("---", test.ID)
}

// 主键id
func TestSaveClauseConflictGitHub(t *testing.T) {
	{
		existOrCreateTable(t, "test_clause_conflict")
	}

	// test 和Columns 都带 id，是会更新的
	test := &TestClauseConflict{
		ID:        28,
		UserId:    100,
		UserName:  "updater" + time.Now().String(),
		CreatedBy: 666,
		CreatedAt: time.Now(),
	}
	err := data.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&TestClauseConflict{}).Debug().Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},

			DoUpdates: clause.AssignmentColumns([]string{
				"created_by", "user_name",
			}),
		}).Create(&test).Error; err != nil {
			return err
		}

		fmt.Println("---", test.ID)

		return nil
	})
	if err != nil {
		t.Fail()
	}
}

// UseIndex 使用
func TestIndex(t *testing.T) {
	{
		existOrCreateTable(t, "test_clause_conflict")
	}

	// test 和Columns 都带 id，是会更新的
	test := &TestClauseConflict{}

	if err := data.db.Model(&TestClauseConflict{}).Debug().Clauses(hints.UseIndex("TEST_CLAUSE_CONFLICT_USER_ID_IDX")).
		Where("user_id=?", 101).First(&test).Error; err != nil {
		t.Log(err)
		t.Fail()
	}

	fmt.Println("---", test)
}
