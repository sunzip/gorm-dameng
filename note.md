# create
来自 go/pkg/mod/github.com/godoes/gorm-dameng@v0.5.0/create.go
增加支持配置Columns，按照 Columns相同时就更新，不同时就新增

使用示例
```
    if err := data.db.Model(&TestClauseConflict{}).Debug().Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "user_name"}},

		DoUpdates: clause.AssignmentColumns([]string{
			"created_by", "created_at",
		}),
	}).Create(&test).Error; err != nil {
		t.Fail()
	}

    DoUpdates 中字段不能和 Columns中字段重复
    test 中的Columns 相关的字段要存在，不存在，则无法判断数据是否是更新或新增
    没有Columns 时，默认是按照主键id做重复判断的，如果test 中id不存在，则为新增
    有Columns 时，如果是id，确保test 中id存在
        因为create默认id是自动生成的，test中如果没有id，则不会在create sql中出现id，会导致生成sql有误。
```