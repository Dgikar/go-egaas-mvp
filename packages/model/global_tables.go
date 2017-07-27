package model

type Tables struct {
	tableName             string
	Name                  []byte `gorm:"primary_key;not null"`
	ColumnsAndPermissions string `gorm:not null;type:jsonb(PostgreSQL)`
	Conditions            string `gorm:not null`
	RbID                  int64  `gorm:not null`
}

func (t *Tables) TableName() string {
	return t.tableName
}

func (t *Tables) SetTableName(tableName string) {
	t.tableName = tableName
}

func (t *Tables) Create() error {
	return DBConn.Create(t).Error
}

func (t *Tables) GetByName(name string) error {
	return DBConn.Where("name = ?", name).First(t).Error
}

func (t *Tables) ExistsByName(name string) (bool, error) {
	query := DBConn.Where("name = ?", name).First(t)
	return !query.RecordNotFound(), query.Error
}

func (t *Tables) GetPermissions(name, jsonKey string) (map[string]string, error) {
	keyStr := ""
	if jsonKey != "" {
		keyStr = `->'` + jsonKey + `'`
	}
	rows, err := DBConn.Raw(`SELECT data.* FROM "`+t.tableName+`", jsonb_each_text(columns_and_permissions`+keyStr+`) AS data WHERE name = ?`, name).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var key, value string
	result := map[string]string{}
	for rows.Next() {
		rows.Scan(&key, &value)
		result[key] = value
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *Tables) SetActionByName(table, name, action, actionValue string, rbID int64) (int64, error) {
	query := DBConn.Exec(`UPDATE "`+table+`" SET columns_and_permissions = jsonb_set(columns_and_permissions, '{`+action+`}', ?, true), rb_id = ? WHERE name = ?`, `"`+actionValue+`"`, rbID, name)
	return query.RowsAffected, query.Error
}

func CreateStateTablesTable(stateID string) error {
	return DBConn.Exec(`CREATE TABLE "` + stateID + `_tables" (
				"name" varchar(100)  NOT NULL DEFAULT '',
				"columns_and_permissions" jsonb,
				"conditions" text  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER TABLE ONLY "` + stateID + `_tables" ADD CONSTRAINT "` + stateID + `_tables_pkey" PRIMARY KEY (name);
	`).Error
}