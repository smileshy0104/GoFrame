package orm

import (
	"database/sql"
	"errors"
	"fmt"
	newLog "frame/log"
	"reflect"
	"strings"
	"time"
)

// FrameDb 代表一个数据库连接和日志记录器的组合，用于对数据库进行操作并记录操作日志。
// 它包含数据库连接实例、日志记录器实例和表名前缀三个组成部分。
type FrameDb struct {
	// db 保存了到数据库的连接实例，通过它来执行SQL查询和操作。
	db *sql.DB

	// logger 用于记录数据库操作的日志，帮助开发者理解数据库操作的上下文。
	logger *newLog.Logger

	// Prefix 是表名的前缀，用于在查询中动态指定表名。
	Prefix string
}

// FrameSession 是一个数据库会话结构体，用于执行数据库操作。
// 它封装了数据库连接、事务处理、查询构建等功能。
type FrameSession struct {
	// db 是数据库连接的实例，用于执行SQL语句。
	db *FrameDb
	// tx 代表当前的数据库事务，如果有的话。
	tx *sql.Tx
	// beginTx 表示是否已经开始了一个事务。
	beginTx bool
	// tableName 是当前操作的数据库表名。
	tableName string
	// fieldName 存储了当前操作涉及的字段名列表。
	fieldName []string
	// placeHolder 用于存储SQL语句中的占位符。
	placeHolder []string
	// values 是与字段名相对应的值，用于插入或更新操作。
	values []any
	// updateParam 用于构建更新操作的SET部分。
	updateParam strings.Builder
	// whereParam 用于构建WHERE子句，以指定更新或查询的条件。
	whereParam strings.Builder
	// whereValues 是WHERE子句中的值，用于匹配条件。
	whereValues []any
}

// Open 是一个用于初始化 FrameDb 数据库连接的方法。
// 它接受数据库驱动名称和数据源作为参数，并返回一个 FrameDb 实例。
// 该方法配置了数据库连接池的各项参数，确保数据库连接的高效和稳定。
func Open(driverName string, source string) *FrameDb {
	// 打开数据库连接
	db, err := sql.Open(driverName, source)
	if err != nil {
		panic(err)
	}
	// 设置最大空闲连接数
	db.SetMaxIdleConns(5)
	// 设置最大连接数
	db.SetMaxOpenConns(100)
	// 设置连接最大存活时间
	db.SetConnMaxLifetime(time.Minute * 3)
	// 设置空闲连接最大存活时间
	db.SetConnMaxIdleTime(time.Minute * 1)

	// 创建 FrameDb 实例
	frameDb := &FrameDb{
		db: db,
		// logger 用于记录数据库操作的日志
		logger: newLog.Default(),
	}
	// 测试连接
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	// 返回 FrameDb 实例
	return frameDb
}

// Close 关闭数据库连接
func (db *FrameDb) Close() error {
	return db.db.Close()
}

// SetMaxIdleConns 最大空闲连接数，默认不配置，是2个最大空闲连接
func (db *FrameDb) SetMaxIdleConns(n int) {
	db.db.SetMaxIdleConns(n)
}

// New 创建一个新的 FrameSession 实例，用于执行数据库操作。
func (db *FrameDb) New(data any) *FrameSession {
	// 创建 FrameSession 实例并将其 db 字段设置为当前 FrameDb 实例。
	m := &FrameSession{
		db: db,
	}

	// 获取 data 参数的类型。
	t := reflect.TypeOf(data)

	// 检查 data 是否为指针类型，如果不是，则抛出 panic。
	// 这是因为后续操作需要通过反射获取指针指向的类型的名称。
	if t.Kind() != reflect.Pointer {
		panic(errors.New("data must be pointer"))
	}

	// 获取指针指向的类型的反射 Type。
	tVar := t.Elem()

	// 如果表名尚未设置，则根据 data 参数指向的类型的名称生成一个表名。
	// 表名由数据库前缀和类型名称的组合而成。
	if m.tableName == "" {
		m.tableName = m.db.Prefix + strings.ToLower(Name(tVar.Name()))
	}

	// 返回初始化后的 FrameSession 实例。
	return m
}

// Table 设置表名
func (s *FrameSession) Table(name string) *FrameSession {
	s.tableName = name
	return s
}

// TODO 重要部分，解析相关的插入数据
// fieldNames 提取数据结构中的字段名和对应值，准备用于SQL查询。
// 该方法主要作用是通过反射机制遍历给定数据结构的字段，根据字段标签确定SQL查询中的字段名和占位符，并将字段值存储起来。
// 参数 data 是一个指向任意类型的指针，该类型将被反射以提取字段信息。
func (s *FrameSession) fieldNames(data any) {
	// 使用反射获取数据的类型和值
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)

	// 确保 data 参数是一个指针类型，以防止反射操作出错
	if t.Kind() != reflect.Pointer {
		panic(errors.New("data must be pointer"))
	}

	// 获取指针所指向的类型的元素类型和值
	tVar := t.Elem()
	vVar := v.Elem()

	// 如果表名尚未设置，则根据数据结构的名称生成一个默认表名
	if s.tableName == "" {
		// 根据数据结构的名称生成表名
		s.tableName = s.db.Prefix + strings.ToLower(Name(tVar.Name()))
	}

	// 遍历数据结构的每个字段
	for i := 0; i < tVar.NumField(); i++ {
		// 获取字段的名称和标签
		fieldName := tVar.Field(i).Name
		tag := tVar.Field(i).Tag

		// 从字段标签中提取gorm标签的值，用于SQL查询
		sqlTag := tag.Get("gorm")
		if sqlTag == "" {
			sqlTag = strings.ToLower(Name(fieldName))
		} else {
			// 如果字段标记包含“auto_increment”，则跳过该字段，因为它通常是自增长的主键
			if strings.Contains(sqlTag, "auto_increment") {
				continue
			}
			// 如果sqlTag包含逗号，取逗号前的部分作为字段名
			if strings.Contains(sqlTag, ",") {
				sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
			}
		}

		// 获取字段的值并进行类型断言
		id := vVar.Field(i).Interface()

		// 如果sqlTag是"id"且字段值是自增长的主键，则跳过
		if strings.ToLower(sqlTag) == "id" && IsAutoId(id) {
			continue
		}

		// 将处理后的字段名和对应的值添加到session的相应切片中
		s.fieldName = append(s.fieldName, sqlTag)
		s.placeHolder = append(s.placeHolder, "?")
		s.values = append(s.values, vVar.Field(i).Interface())
	}
}

// TODO 重要部分，解析相关的插入数据 根据字段的 tag 决定是否将值添加到 s.values 中。
// batchValues 是一个用于处理批量插入数据的函数。
// 它接受一个 any 类型的切片 data，该切片包含了多个结构体对象。
// 函数会遍历每个结构体对象，提取其字段值，并根据字段的 tag 决定是否将值添加到 s.values 中。
// s.values 是一个用于存储所有数据的切片，这些数据将用于数据库的批量插入操作。
func (s *FrameSession) batchValues(data []any) {
	// 初始化 s.values 为一个新的空切片，用于存储处理后的字段值。
	s.values = make([]any, 0)

	// 遍历 data 切片中的每个元素。
	for _, v := range data {
		// 获取当前元素的类型和值。
		t := reflect.TypeOf(v)
		v := reflect.ValueOf(v)

		// 检查元素是否为指针类型，如果不是，则抛出错误。
		if t.Kind() != reflect.Pointer {
			panic(errors.New("data must be pointer"))
		}

		// 获取指针所指向的类型的变量和值。
		tVar := t.Elem()
		vVar := v.Elem()

		// 遍历类型的每个字段。
		for i := 0; i < tVar.NumField(); i++ {
			// 获取当前字段的名称和 tag。
			fieldName := tVar.Field(i).Name
			tag := tVar.Field(i).Tag
			sqlTag := tag.Get("gorm")

			// 如果 gorm tag 为空，则使用字段名称的小写形式作为默认值。
			if sqlTag == "" {
				sqlTag = strings.ToLower(Name(fieldName))
			} else {
				// 如果 tag 中包含 "auto_increment"，表示该字段是自增长的主键，跳过该字段。
				if strings.Contains(sqlTag, "auto_increment") {
					continue
				}
			}

			// 获取字段的值并进行类型断言。
			id := vVar.Field(i).Interface()

			// 如果字段的 tag 是 "id" 且值是自动生成的 ID，则跳过该字段。
			if strings.ToLower(sqlTag) == "id" && IsAutoId(id) {
				continue
			}

			// 将字段的值添加到 s.values 中。
			s.values = append(s.values, vVar.Field(i).Interface())
		}
	}
}

// Insert 方法用于向数据库中插入一条记录。
// 参数 data 代表要插入的数据，其类型为任意类型。
// 返回值为插入记录的自增ID、受影响的行数以及可能的错误。
// 该方法首先构建插入SQL语句，然后根据是否在事务中选择不同的数据库连接进行预编译。
// 预编译成功后执行SQL语句，并获取执行结果。
// 最后，从执行结果中获取最后插入记录的自增ID和受影响的行数，并返回这些值。
func (s *FrameSession) Insert(data any) (int64, int64, error) {
	// 构建插入SQL语句的字段名部分。（解析相关的插入数据）
	s.fieldNames(data)
	// 构建完整的插入SQL语句。
	query := fmt.Sprintf("insert into %s (%s) values (%s)", s.tableName, strings.Join(s.fieldName, ","), strings.Join(s.placeHolder, ","))
	// 记录SQL语句日志。
	s.db.logger.Info(query)

	// 声明stmt变量用于存储预编译的SQL语句。
	var stmt *sql.Stmt
	// 声明err变量用于存储错误信息。
	var err error

	// 根据是否在事务中选择不同的数据库连接进行预编译。
	if s.beginTx {
		stmt, err = s.tx.Prepare(query)
	} else {
		stmt, err = s.db.db.Prepare(query)
	}
	// 如果预编译失败，返回错误。
	if err != nil {
		return -1, -1, err
	}

	// 执行预编译的SQL语句。
	r, err := stmt.Exec(s.values...)
	// 如果执行失败，返回错误。
	if err != nil {
		return -1, -1, err
	}

	// 获取最后插入记录的自增ID。
	id, err := r.LastInsertId()
	// 如果获取失败，返回错误。
	if err != nil {
		return -1, -1, err
	}

	// 获取受影响的行数。
	affected, err := r.RowsAffected()
	// 如果获取失败，返回错误。
	if err != nil {
		return -1, -1, err
	}

	// 返回插入记录的自增ID、受影响的行数以及nil错误。
	return id, affected, nil
}

// InsertBatch 批量插入数据到数据库中。
// 该方法根据提供的数据数组生成一个批量插入查询，并执行该查询。
func (s *FrameSession) InsertBatch(data []any) (int64, int64, error) {
	// 当数据为空时，返回错误。
	if len(data) == 0 {
		return -1, -1, errors.New("no data insert")
	}

	// 准备插入查询的字段名。（通过第一个数据获取对应信息）
	s.fieldNames(data[0])

	// 构建插入查询的初始部分，包括表名和字段名。
	query := fmt.Sprintf("insert into %s (%s) values ", s.tableName, strings.Join(s.fieldName, ","))

	// 构建包含多个值集合的字符串，每个值集合代表一行数据。（拼接成批量插入的sql语句）
	var sb strings.Builder
	sb.WriteString(query)
	for index, _ := range data {
		sb.WriteString("(")
		sb.WriteString(strings.Join(s.placeHolder, ","))
		sb.WriteString(")")
		if index < len(data)-1 {
			sb.WriteString(",")
		}
	}

	// 将所有数据记录的值添加到batchValues中，以备后续执行查询。
	s.batchValues(data)

	// 记录日志信息。
	s.db.logger.Info(sb.String())

	// 准备SQL语句。
	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(sb.String())
	} else {
		stmt, err = s.db.db.Prepare(sb.String())
	}

	// 如果准备SQL语句时发生错误，返回错误。
	if err != nil {
		return -1, -1, err
	}

	// 执行SQL语句。
	r, err := stmt.Exec(s.values...)
	if err != nil {
		return -1, -1, err
	}

	// 获取最后插入行的ID。
	id, err := r.LastInsertId()
	if err != nil {
		return -1, -1, err
	}

	// 获取受影响的行数。
	affected, err := r.RowsAffected()
	if err != nil {
		return -1, -1, err
	}

	// 返回最后插入行的ID、受影响的行数和nil错误。
	return id, affected, nil
}

// UpdateParam 更新FrameSession对象中的参数。
// 该方法用于动态构建SQL更新语句的SET部分，通过接受字段名称和对应的值来更新session的状态。
func (s *FrameSession) UpdateParam(field string, value any) *FrameSession {
	// 检查是否已经有参数被添加，如果有，则添加逗号分隔。
	if s.updateParam.String() != "" {
		s.updateParam.WriteString(",")
	}
	// 将字段名称和对应的占位符添加到updateParam中，用于后续构建SQL语句。
	s.updateParam.WriteString(field)
	s.updateParam.WriteString(" = ? ")
	// 将实际的值添加到values切片中，用于后续的SQL查询。
	s.values = append(s.values, value)
	// 返回FrameSession对象，支持链式调用。
	return s
}

// UpdateMap 更新FrameSession对象的状态，以便在后续的SQL语句中使用。
// 该方法接受一个键值对映射（data），将其转换为SQL的SET子句需要的格式。
func (s *FrameSession) UpdateMap(data map[string]any) *FrameSession {
	// 遍历data映射，构建SQL的SET子句需要的部分。
	for k, v := range data {
		// 如果已经有字段需要更新，添加逗号分隔。
		if s.updateParam.String() != "" {
			s.updateParam.WriteString(",")
		}
		// 将字段名和对应的占位符"?"添加到updateParam中。
		s.updateParam.WriteString(k)
		s.updateParam.WriteString(" = ? ")
		// 将字段的值添加到values切片中，作为SQL语句的参数。
		s.values = append(s.values, v)
	}
	// 返回FrameSession对象，支持链式调用。
	return s
}

// Update 更新数据库中的记录。
// 该方法支持两种调用方式：
// 1. 传递键值对（列名和新值），例如：Update("age", 1)
// 2. 传递一个结构体指针，例如：Update(user)
// 参数说明：
//   - data: 可变参数，用于指定更新的字段或结构体。如果传递两个参数，则第一个参数为列名，第二个参数为新值；
//     如果传递一个参数，则该参数应为一个结构体指针。
func (s *FrameSession) Update(data ...any) (int64, int64, error) {
	// 检查参数数量是否合法。如果参数数量超过2个，则返回错误。
	if len(data) > 2 {
		return -1, -1, errors.New("param not valid")
	}

	// 如果没有传递任何参数，则执行无条件更新操作。
	if len(data) == 0 {
		// 构建更新SQL语句。
		query := fmt.Sprintf("update %s set %s", s.tableName, s.updateParam.String())
		var sb strings.Builder
		sb.WriteString(query)
		sb.WriteString(s.whereParam.String())
		s.db.logger.Info(sb.String())

		// 根据事务状态选择不同的数据库连接进行预编译。
		var stmt *sql.Stmt
		var err error
		if s.beginTx {
			stmt, err = s.tx.Prepare(sb.String())
		} else {
			stmt, err = s.db.db.Prepare(sb.String())
		}
		if err != nil {
			return -1, -1, err
		}

		// 执行SQL语句并获取结果。
		s.values = append(s.values, s.whereValues...)
		r, err := stmt.Exec(s.values...)
		if err != nil {
			return -1, -1, err
		}
		id, err := r.LastInsertId()
		if err != nil {
			return -1, -1, err
		}
		affected, err := r.RowsAffected()
		if err != nil {
			return -1, -1, err
		}
		return id, affected, nil
	}

	// 判断是单个结构体还是键值对更新。
	single := true
	if len(data) == 2 {
		single = false
	}

	// 如果是键值对更新，则直接构建SET子句。
	if !single {
		if s.updateParam.String() != "" {
			s.updateParam.WriteString(",")
		}
		s.updateParam.WriteString(data[0].(string))
		s.updateParam.WriteString(" = ? ")
		s.values = append(s.values, data[1])
	} else {
		// 如果是结构体更新，则通过反射提取结构体字段信息。
		updateData := data[0]
		t := reflect.TypeOf(updateData)
		v := reflect.ValueOf(updateData)

		// 确保传递的是一个指针类型。
		if t.Kind() != reflect.Pointer {
			panic(errors.New("updateData must be pointer"))
		}

		tVar := t.Elem()
		vVar := v.Elem()

		// 遍历结构体字段，构建SET子句。
		for i := 0; i < tVar.NumField(); i++ {
			fieldName := tVar.Field(i).Name
			tag := tVar.Field(i).Tag
			sqlTag := tag.Get("gorm")

			// 如果字段标记包含"auto_increment"，则跳过该字段。
			if sqlTag == "" {
				sqlTag = strings.ToLower(Name(fieldName))
			} else {
				if strings.Contains(sqlTag, "auto_increment") {
					continue
				}
				if strings.Contains(sqlTag, ",") {
					sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
				}
			}

			id := vVar.Field(i).Interface()

			// 如果字段是自增长主键且值为默认值，则跳过该字段。
			if strings.ToLower(sqlTag) == "id" && IsAutoId(id) {
				continue
			}

			if s.updateParam.String() != "" {
				s.updateParam.WriteString(",")
			}
			s.updateParam.WriteString(sqlTag)
			s.updateParam.WriteString(" = ? ")
			s.values = append(s.values, vVar.Field(i).Interface())
		}
	}

	// 构建最终的更新SQL语句。
	query := fmt.Sprintf("update %s set %s", s.tableName, s.updateParam.String())
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParam.String())
	s.db.logger.Info(sb.String())

	// 根据事务状态选择不同的数据库连接进行预编译。
	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(sb.String())
	} else {
		stmt, err = s.db.db.Prepare(sb.String())
	}
	if err != nil {
		return -1, -1, err
	}

	// 执行SQL语句并获取结果。
	s.values = append(s.values, s.whereValues...)
	r, err := stmt.Exec(s.values...)
	if err != nil {
		return -1, -1, err
	}
	id, err := r.LastInsertId()
	if err != nil {
		return -1, -1, err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return -1, -1, err
	}
	return id, affected, nil
}

func (s *FrameSession) Delete() (int64, error) {
	//delete from table where id=?
	query := fmt.Sprintf("delete from %s ", s.tableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParam.String())
	s.db.logger.Info(sb.String())

	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(sb.String())
	} else {
		stmt, err = s.db.db.Prepare(sb.String())
	}
	if err != nil {
		return 0, err
	}
	r, err := stmt.Exec(s.whereValues...)
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}

func (s *FrameSession) Select(data any, fields ...string) ([]any, error) {
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Pointer {
		return nil, errors.New("data must be pointer")
	}
	fieldStr := "*"
	if len(fields) > 0 {
		fieldStr = strings.Join(fields, ",")
	}
	query := fmt.Sprintf("select %s from %s ", fieldStr, s.tableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParam.String())
	s.db.logger.Info(sb.String())

	stmt, err := s.db.db.Prepare(sb.String())
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(s.whereValues...)
	if err != nil {
		return nil, err
	}
	//id user_name age
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	result := make([]any, 0)
	for {
		if rows.Next() {
			//由于 传进来的是一个指针地址 如果每次赋值，实际都是一个 result里面 值都一样
			//每次查询的时候 data都重新换一个地址
			data := reflect.New(t.Elem()).Interface()
			values := make([]any, len(columns))
			fieldScan := make([]any, len(columns))
			for i := range fieldScan {
				fieldScan[i] = &values[i]
			}
			err := rows.Scan(fieldScan...)
			if err != nil {
				return nil, err
			}
			tVar := t.Elem()
			vVar := reflect.ValueOf(data).Elem()
			for i := 0; i < tVar.NumField(); i++ {
				name := tVar.Field(i).Name
				tag := tVar.Field(i).Tag
				//id,auto
				sqlTag := tag.Get("msorm")
				if sqlTag == "" {
					sqlTag = strings.ToLower(Name(name))
				} else {
					if strings.Contains(sqlTag, ",") {
						sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
					}
				}

				for j, colName := range columns {
					if sqlTag == colName {
						target := values[j]
						targetValue := reflect.ValueOf(target)
						fieldType := tVar.Field(i).Type
						//这样不行 类型不匹配 转换类型
						result := reflect.ValueOf(targetValue.Interface()).Convert(fieldType)
						vVar.Field(i).Set(result)
					}
				}

			}
			result = append(result, data)
		} else {
			break
		}
	}

	return result, nil
}

func (s *FrameSession) SelectOne(data any, fields ...string) error {
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Pointer {
		return errors.New("data must be pointer")
	}
	fieldStr := "*"
	if len(fields) > 0 {
		fieldStr = strings.Join(fields, ",")
	}
	query := fmt.Sprintf("select %s from %s ", fieldStr, s.tableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParam.String())
	s.db.logger.Info(sb.String())

	stmt, err := s.db.db.Prepare(sb.String())
	if err != nil {
		return err
	}
	rows, err := stmt.Query(s.whereValues...)
	if err != nil {
		return err
	}
	//id user_name age
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]any, len(columns))
	fieldScan := make([]any, len(columns))
	for i := range fieldScan {
		fieldScan[i] = &values[i]
	}
	if rows.Next() {
		err := rows.Scan(fieldScan...)
		if err != nil {
			return err
		}
		tVar := t.Elem()
		vVar := reflect.ValueOf(data).Elem()
		for i := 0; i < tVar.NumField(); i++ {
			name := tVar.Field(i).Name
			tag := tVar.Field(i).Tag
			//id,auto
			sqlTag := tag.Get("msorm")
			if sqlTag == "" {
				sqlTag = strings.ToLower(Name(name))
			} else {
				if strings.Contains(sqlTag, ",") {
					sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
				}
			}

			for j, colName := range columns {
				if sqlTag == colName {
					target := values[j]
					targetValue := reflect.ValueOf(target)
					fieldType := tVar.Field(i).Type
					//这样不行 类型不匹配 转换类型
					result := reflect.ValueOf(targetValue.Interface()).Convert(fieldType)
					vVar.Field(i).Set(result)
				}
			}

		}
	}
	return nil
}

func (s *FrameSession) Count() (int64, error) {
	return s.Aggregate("count", "*")
}

func (s *FrameSession) Aggregate(funcName string, field string) (int64, error) {
	var fieldSb strings.Builder
	fieldSb.WriteString(funcName)
	fieldSb.WriteString("(")
	fieldSb.WriteString(field)
	fieldSb.WriteString(")")
	query := fmt.Sprintf("select %s from %s ", fieldSb.String(), s.tableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(s.whereParam.String())
	s.db.logger.Info(sb.String())

	stmt, err := s.db.db.Prepare(sb.String())
	if err != nil {
		return 0, err
	}
	row := stmt.QueryRow(s.whereValues...)
	if row.Err() != nil {
		return 0, err
	}
	var result int64
	err = row.Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

// 原生sql的支持
func (s *FrameSession) Exec(query string, values ...any) (int64, error) {
	var stmt *sql.Stmt
	var err error
	if s.beginTx {
		stmt, err = s.tx.Prepare(query)
	} else {
		stmt, err = s.db.db.Prepare(query)
	}
	if err != nil {
		return 0, err
	}
	r, err := stmt.Exec(values)
	if err != nil {
		return 0, err
	}
	if strings.Contains(strings.ToLower(query), "insert") {
		return r.LastInsertId()
	}
	return r.RowsAffected()
}

func (s *FrameSession) QueryRow(sql string, data any, queryValues ...any) error {
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Pointer {
		return errors.New("data must be pointer")
	}
	stmt, err := s.db.db.Prepare(sql)
	if err != nil {
		return err
	}
	rows, err := stmt.Query(queryValues...)
	if err != nil {
		return err
	}
	//id user_name age
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]any, len(columns))
	fieldScan := make([]any, len(columns))
	for i := range fieldScan {
		fieldScan[i] = &values[i]
	}
	if rows.Next() {
		err := rows.Scan(fieldScan...)
		if err != nil {
			return err
		}
		tVar := t.Elem()
		vVar := reflect.ValueOf(data).Elem()
		for i := 0; i < tVar.NumField(); i++ {
			name := tVar.Field(i).Name
			tag := tVar.Field(i).Tag
			//id,auto
			sqlTag := tag.Get("msorm")
			if sqlTag == "" {
				sqlTag = strings.ToLower(Name(name))
			} else {
				if strings.Contains(sqlTag, ",") {
					sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
				}
			}

			for j, colName := range columns {
				if sqlTag == colName {
					target := values[j]
					targetValue := reflect.ValueOf(target)
					fieldType := tVar.Field(i).Type
					//这样不行 类型不匹配 转换类型
					result := reflect.ValueOf(targetValue.Interface()).Convert(fieldType)
					vVar.Field(i).Set(result)
				}
			}

		}
	}
	return nil
}

// Begin 开始一个新的事务。
// 返回错误如果数据库操作失败。
func (s *FrameSession) Begin() error {
	// 获取sql.DB中的事务
	tx, err := s.db.db.Begin()
	if err != nil {
		return err
	}
	// 设置事务为true
	s.tx = tx
	s.beginTx = true
	return nil
}

// Commit 提交当前事务。
// 返回错误如果数据库操作失败。
func (s *FrameSession) Commit() error {
	// 提交事务
	err := s.tx.Commit()
	if err != nil {
		return err
	}
	s.beginTx = false
	return nil
}

// Rollback 回滚当前事务。
// 返回错误如果数据库操作失败。
func (s *FrameSession) Rollback() error {
	// 回滚事务
	err := s.tx.Rollback()
	if err != nil {
		return err
	}
	s.beginTx = false
	return nil
}

// Where 为查询添加一个等值条件。
// 参数 field 是列名，value 是对应的值。
// 返回修改后的 FrameSession 实例。
func (s *FrameSession) Where(field string, value any) *FrameSession {
	//id=1 and name=xx
	if s.whereParam.String() == "" {
		s.whereParam.WriteString(" where ")
	}
	s.whereParam.WriteString(field)
	s.whereParam.WriteString(" = ")
	s.whereParam.WriteString(" ? ")
	s.whereValues = append(s.whereValues, value)
	return s
}

// Like 为查询添加一个模糊条件（右侧包含）。
// 参数 field 是列名，value 是对应的值。
// 返回修改后的 FrameSession 实例。
func (s *FrameSession) Like(field string, value any) *FrameSession {
	//name like %s%
	if s.whereParam.String() == "" {
		s.whereParam.WriteString(" where ")
	}
	s.whereParam.WriteString(field)
	s.whereParam.WriteString(" like ")
	s.whereParam.WriteString(" ? ")
	s.whereValues = append(s.whereValues, "%"+value.(string)+"%")
	return s
}

// LikeRight 为查询添加一个右侧模糊条件。
// 参数 field 是列名，value 是对应的值。
// 返回修改后的 FrameSession 实例。
func (s *FrameSession) LikeRight(field string, value any) *FrameSession {
	//name like %s%
	if s.whereParam.String() == "" {
		s.whereParam.WriteString(" where ")
	}
	s.whereParam.WriteString(field)
	s.whereParam.WriteString(" like ")
	s.whereParam.WriteString(" ? ")
	s.whereValues = append(s.whereValues, value.(string)+"%")
	return s
}

// LikeLeft 为查询添加一个左侧模糊条件。
// 参数 field 是列名，value 是对应的值。
// 返回修改后的 FrameSession 实例。
func (s *FrameSession) LikeLeft(field string, value any) *FrameSession {
	//name like %s%
	if s.whereParam.String() == "" {
		s.whereParam.WriteString(" where ")
	}
	s.whereParam.WriteString(field)
	s.whereParam.WriteString(" like ")
	s.whereParam.WriteString(" ? ")
	s.whereValues = append(s.whereValues, "%"+value.(string))
	return s
}

// Group 为查询添加一个分组条件。
// 参数 field 是列名列表，用于分组。
// 返回修改后的 FrameSession 实例。
func (s *FrameSession) Group(field ...string) *FrameSession {
	//group by aa,bb
	s.whereParam.WriteString(" group by ")
	s.whereParam.WriteString(strings.Join(field, ","))
	return s
}

// OrderDesc 为查询添加一个降序排序条件。
// 参数 field 是列名列表，用于排序。
// 返回修改后的 FrameSession 实例。
func (s *FrameSession) OrderDesc(field ...string) *FrameSession {
	//order by aa,bb desc
	s.whereParam.WriteString(" order by ")
	s.whereParam.WriteString(strings.Join(field, ","))
	s.whereParam.WriteString(" desc ")
	return s
}

// OrderAsc 为查询添加一个升序排序条件。
// 参数 field 是列名列表，用于排序。
// 返回修改后的 FrameSession 实例。
func (s *FrameSession) OrderAsc(field ...string) *FrameSession {
	//order by aa,bb asc
	s.whereParam.WriteString(" order by ")
	s.whereParam.WriteString(strings.Join(field, ","))
	s.whereParam.WriteString(" asc ")
	return s
}

// Order 为查询添加一个自定义排序条件。
// 参数 field 是列名和排序方式的交替列表。
// 返回修改后的 FrameSession 实例。
// 如果列名数量不是偶数，抛出 panic。
func (s *FrameSession) Order(field ...string) *FrameSession {
	if len(field)%2 != 0 {
		panic("field num not true")
	}
	s.whereParam.WriteString(" order by ")
	for index, v := range field {
		s.whereParam.WriteString(v + " ")
		if index%2 != 0 && index < len(field)-1 {
			s.whereParam.WriteString(",")
		}
	}
	return s
}

// And 在对应的条件后面添加 and
func (s *FrameSession) And() *FrameSession {
	s.whereParam.WriteString(" and ")
	return s
}

// Or 在对应的条件后面添加 or
func (s *FrameSession) Or() *FrameSession {
	s.whereParam.WriteString(" or ")
	return s
}

// IsAutoId 判断给定的id是否为自动增长的ID。
// 自动增长的ID定义为非负的整数类型（int, int32, int64）且值大于0。
// 如果id符合自动增长ID的定义，则返回true，否则返回false。
func IsAutoId(id any) bool {
	// 获取id的类型
	t := reflect.TypeOf(id)
	// 根据id的类型进行判断
	switch t.Kind() {
	case reflect.Int64:
		// 如果是int64类型且值小于等于0，则认为是自动增长的ID
		if id.(int64) <= 0 {
			return true
		}
	case reflect.Int32:
		// 如果是int32类型且值小于等于0，则认为是自动增长的ID
		if id.(int32) <= 0 {
			return true
		}
	case reflect.Int:
		// 如果是int类型且值小于等于0，则认为是自动增长的ID
		if id.(int) <= 0 {
			return true
		}
	default:
		// 如果不是上述支持的整数类型，则不认为是自动增长的ID
		return false
	}
	// 如果不符合自动增长ID的条件，则返回false
	return false
}

// Name 函数用于将驼峰命名转换为下划线命名。
// 参数 name 是一个字符串，表示驼峰命名的单词。
// 返回值是一个字符串，表示转换后的下划线命名的单词。
func Name(name string) string {
	// 创建一个字符串切片来存储输入的字符串
	var names = name[:]

	// 初始化 lastIndex 为 0，用于记录最后一个大写字母的位置
	lastIndex := 0

	// 创建一个 strings.Builder 对象，用于构建最终的字符串
	var sb strings.Builder

	// 遍历 names 中的每个字符及其索引
	for index, value := range names {
		// 检查字符是否为大写字母
		if value >= 65 && value <= 90 {
			// 如果当前字符是大写字母且不是第一个字符，则在它之前添加下划线
			if index == 0 {
				continue
			}
			// 向 sb 中添加从 lastIndex 到当前索引的子字符串
			sb.WriteString(name[:index])
			// 添加下划线
			sb.WriteString("_")
			// 更新 lastIndex 为当前索引
			lastIndex = index
		}
	}
	// 向 sb 中添加从 lastIndex 到字符串末尾的子字符串
	sb.WriteString(name[lastIndex:])
	// 返回构建好的字符串
	return sb.String()
}
