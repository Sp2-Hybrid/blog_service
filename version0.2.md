### Version0.2

在之前的工作中，我们已经完成了项目的前期准备工作，接下来的Version0.2是项目的基本使用部分。

这部分分为三个主要内容：

- 模块开发
- 文件服务
- 访问控制

## 2.1 模块开发

| 功能         | HTTP 方法 | 路径      |
| ------------ | --------- | --------- |
| 新增标签     | POST      | /tags     |
| 删除指定标签 | DELETE    | /tags/:id |
| 更新指定标签 | PUT       | /tags/:id |
| 获取标签列表 | GET       | /tags     |

对于标签管理有上表所示。

### 2.1.1 标签模块模型操作

首先需要对标签表进行处理，在项目的`internal/model`目录下新建tag.go文件，针对标签模块的模型操作进行封装，并且只与实体产生关系。

```go
func (t Tag)Count(db *gorm.DB)(int, error){
    var count int
    if t.Name!=""{
        db = db.Where("name = ?", t.Name)
    }
    db = db.Where("state = ?", t.State)
    if err := db.Model(&t).Where("is_del = ?", 0).Count(&count).Error; err!=nil{
        return 0, err
    }
    return count, nil
}

func (t Tag)List(db *gorm.DB, pageOffset, pageSize int)([]*Tag, err){
    var tags []*Tag
    var err error 
    if pageOffset >= 0 && pageSize > 0{
        db = db.Offset(pageOffset).Limit(pageSize)
    }
    if t.Name != ""{
        db = db.Where("name = ?", t.Name)
    }
    db = db.Where("state = ?", t.State)
    if err = db.Where("is_del = ?", 0).Find(&tags).Error; err!=nil{
        return nil, err
    }
    return tags, nil
}

func (t Tag)ListByIDs(db *gorm.DB, ids []uint32)([]*Tags, error){
    var tags []*Tag
    var err error 
    db = db.Where("state = ? AND is_del = ?", t.State, 0)
    err = db.Where("id IN (?)", ids).Find(&tags).Error
    if err!=nil && err!=gorm.ErrRecordNotFound{
        return nil, err
    }
    return tags, nil
}

func (t Tag)Get(db *gorm.DB)(Tag, err){
    var tag Tag
    err := db.Where("id = ? AND is_del = ? and state = ?", t.Model.ID, 0, t.State).First(&tag).Error
    if err!=nil && err!=gorm.ErrRecordNotFound{
        return nil, err
    }
    return tag, nil
}

func (t Tag)Create(db *gorm.DB)error{
    return db.Create(&t).Error
}

func (t Tag)Update(db *gorm.DB, value interface{})error{
    return db.Model(&t).Where("id = ? AND is_del = ?", t.Model.ID, 0).Updates(values).Error
}

func (t Tag)Delete(db *gorm.DB) error{
    return db.Where("id = ? AND is_del = ?", t.Model.ID, 0).Delete(&t).Error
}
```

### 2.1.2 处理model回调

我们在编写model代码时，并没有针对公共字段`create_on`、`modified_on`、`deleted_on`、`is_del`进行处理，像这些公用的字段，我们应当尽量简化代码，因此可以使用回调来处理。

分别进行如下的回调相关行为：

- 注册一个新的回调
- 删除现有的回调
- 替换现有的回调
- 注册回调的先后顺序

**新增行为回调**

```go
func updateTimeStampForCreateCallback(scope *gorm.Scope){
    if !scope.HasError(){
        nowTime := time.Now().Unix()
        if createTimeField, ok := scope.FieldByName("CreatedOn"); ok{
            if createTimeField.IsBlank{
                _ = createTimeField.Set(nowTime)
            }
        }
        if modifyTimeField, ok := scope.FieldByName("modifiedOn"); ok{
            if modifyTimeField.IsBlank{
                _ = modifyTimeField.Set(nowTime)
            }
        }
    }
}
```

- 通过调用`scope.FieldByName`方法，获取当前是否包含所需的字段
- 通过判断`Field.IsBlank`的值，可以得知该字段的值是否为空
- 若为空，则会调用`Field.Set`方法给该字段设置值，入参类型为`interface{}`，内部也就是通过反射进行一系列操作赋值

**更新行为的回调**

```go
func updateTimeStampForUpdateCallback(scope *gorm.Scope){
    if _, ok := scope.Get("gorm:update_column"); !ok{
        _ = scope.SetColumn("ModifiedOn", time.Now().Unix())
    }
}
```

- 通过调用`scope.Get("gorm:update_column")`去获取当前设置了标识`gorm:update_column`的字段属性
- 若不存在，也就是没有自定义设置`update_column`，那么将会在更新回调内设置默认字段，`ModifiedOn`的值为当前的时间戳。

**删除行为回调**



**注册行为回调**



----

### 2.1.3 新建dao方法

在项目的`internal/dao`目录下新建dao.go文件，写入如下代码：

```go
type Dao struct{
    engine *gorm.DB
}

func New(engine *gorm.DB) *Dao{
    return &Dao(engine: engine)
}
```

接下来在同层级下新建tag.go文件，用于处理标签模块的dao操作，写入如下代码：

```go
func (d *Dao)CountTag(name string, state uint8) (int, error){
	tag := model.Tag{Name:name, State:state}
	return tag.Count(d.engine)
}

func (d *Dao)GetTagList(name string, state uint8, page, pageSize int)([]*model.Tag, error)  {
	tag := model.Tag{Name:name, State:state}
	pageOffset := app.GetPageOffset(page, pageSize)
	return tag.List(d.engine, pageOffset, pageSize)
}

func (d *Dao)CreatTag(name string, state uint8, createdBy string)error  {
	tag := model.Tag{Name:name, State:state, Model:&(model.Model{CreatedBy:createdBy})}
	return tag.Create(d.engine)
}

func (d *Dao)UpdateTag(id uint32, name string, state uint8, modifiedBy string)error{
	tag := model.Tag{
		Model: &(model.Model{ModifiedBy:modifiedBy, ID:id}),
		Name:  name,
		State: state,
	}
	return tag.Update(d.engine)
}

func (d *Dao)DeleteTag(id uint32) error{
	tag := model.Tag{
		Model: &(model.Model{ID:id}),
	}
	return tag.Delete(d.engine)
}
```

----

其实到此为止我们已经完成了模块的开发，接下来要在上面的基础上再包装一层，完成接口参数的校验工作。

## 2.2 接口参数校验

这部分还可以参考李文周的博客，讲解的很好：https://www.liwenzhou.com/posts/Go/validator_usages/

在这部分我们使用的是validator

**安装**

```go
go get -u github.com/go-playground/validator/v10
```

### 2.2.1 结构体校验

在`internal/service`目录下的`tag.go`文件中，**针对入参校验增加绑定/验证结构体，在路由方法前写入如下代码：**

```go
type CountTagRequest struct {
	Name string `form:"name" binding:"max=100"`
	State uint8 `form:"state,default=1" binding:"oneof=0 1"`
}

type TagListRequest struct{
	Name string `form:"name" binding:"max=100"`
	State uint8 `form:"state,default=1" binding:"oneof=0 1"`
}

type CreateTagRequest struct {
	Name string 	`form:"name" binding:"required,min=2,max=100"`
	CreatedBy 	string 	`form:"created_by" binding:"required,min=2,max=100"`
	State 	uint8 	`form:"state,default=1" binding:"oneof=0 1"`
}

type UpdateTagRequest struct {
	ID 	uint32 		`form:"id" binding:"required,gte=1"`
	Name 	string 	`form:"name" binding:"max=100"`
	State 	uint8 	`form:"state" binding:"oneof=0 1"`
	ModifiedBy 	string 	`form:"modified_by" binding:"required,min=2,max=100"`
}

type DeleteTagRequest struct {
	ID uint32 `form:"id" binding:"required,get=1"`
}
```

### 2.2.2 国际化

完成了上面结构体之后，我们在绑定时已经可以进行初步的参数验证了，但是会存在一个问题，就是输出的错误信息是英文，当我们想要输出中文的报错信息时，需要进行国际化操作。

```go
import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	"github.com/go-playground/locales/zh_Hant_TW"
	ut "github.com/go-playground/universal-translator"
	validator "github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

// 作为中间件使用
func Translations() gin.HandlerFunc {
	return func(c *gin.Context) {
        // zh.New() 中文翻译器
        // en.New() 英文翻译器
		uni := ut.New(en.New(), zh.New(), zh_Hant_TW.New())
		locale := c.GetHeader("locale")
		trans, _ := uni.GetTranslator(locale)
        
        // 修改gin框架中的Validator引擎属性，实现自定制
		v, ok := binding.Validator.Engine().(*validator.Validate)
		if ok {
			switch locale {
			case "zh":
				_ = zh_translations.RegisterDefaultTranslations(v, trans)
				break
			case "en":
				_ = en_translations.RegisterDefaultTranslations(v, trans)
				break
			default:
				_ = zh_translations.RegisterDefaultTranslations(v, trans)
				break
			}
			c.Set("trans", trans)
		}

		c.Next()
	}
}
```

上面的国际化可以进一步作为中间件，读取上下文，便于后续使用

### 2.2.3 注册中间件

回到项目的 `internal/routers` 目录下的 router.go 文件，新增中间件 Translations 的注册，新增代码如下：

```go
func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.Translations())
	...
}
```

### 2.2.4 接口校验

在项目下的 `pkg/app` 目录新建 form.go 文件，写入如下代码：

```go
import (
	...
	ut "github.com/go-playground/universal-translator"
	val "github.com/go-playground/validator/v10"
)

type ValidError struct {
	Key     string
	Message string
}

type ValidErrors []*ValidError

func (v *ValidError) Error() string {
	return v.Message
}

func (v ValidErrors) Error() string {
	return strings.Join(v.Errors(), ",")
}

func (v ValidErrors) Errors() []string {
	var errs []string
	for _, err := range v {
		errs = append(errs, err.Error())
	}

	return errs
}

func BindAndValid(c *gin.Context, v interface{}) (bool, ValidErrors) {
	var errs ValidErrors
	err := c.ShouldBind(v)
	if err != nil {
		v := c.Value("trans")
		trans, _ := v.(ut.Translator)
		verrs, ok := err.(val.ValidationErrors)
		if !ok {
			return false, errs
		}

		for key, value := range verrs.Translate(trans) {
			errs = append(errs, &ValidError{
				Key:     key,
				Message: value,
			})
		}

		return false, errs
	}

	return true, nil
}
```

在上述代码中，我们主要是针对入参校验的方法进行了二次封装，在 BindAndValid 方法中，通过 ShouldBind 进行参数绑定和入参校验，当发生错误后，再通过上一步在中间件 Translations 设置的 Translator 来对错误消息体进行具体的翻译行为。

### 2.2.5 验证

我们回到项目的 `internal/routers/api/v1` 下的 tag.go 文件，修改获取多个标签的 List 接口，用于验证 validator 是否正常，修改代码如下：

```go
func (t Tag) List(c *gin.Context) {
	param := struct {
		Name  string `form:"name" binding:"max=100"`
		State uint8  `form:"state,default=1" binding:"oneof=0 1"`
	}{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, &param)
	if !valid {
		global.Logger.Errorf("app.BindAndValid errs: %v", errs)
		response.ToErrorResponse(errcode.InvalidParams.WithDetails(errs.Errors()...))
		return
	}

	response.ToResponse(gin.H{})
	return
}
```

----

## 2.3 模块开发+

在完成了接口参数校验的学习后，我们可以进一步封装我们前面实现的函数，来实现接口校验的功能。

### 2.3.1 新建service方法

在`internal/service`目录下的tag.go文件中添加如下内容：

```go
func (svc *Service) CountTag(param *CountTagRequest) (int, error) {
	return svc.dao.CountTag(param.Name, param.State)
}

func (svc *Service) GetTagList(param *TagListRequest, pager *app.Pager) ([]*model.Tag, error) {
	return svc.dao.GetTagList(param.Name, param.State, pager.Page, pager.PageSize)
}

func (svc *Service) CreateTag(param *CreateTagRequest) error {
	return svc.dao.CreateTag(param.Name, param.State, param.CreatedBy)
}

func (svc *Service) UpdateTag(param *UpdateTagRequest) error {
	return svc.dao.UpdateTag(param.ID, param.Name, param.State, param.ModifiedBy)
}

func (svc *Service) DeleteTag(param *DeleteTagRequest) error {
	return svc.dao.DeleteTag(param.ID)
}
```

在上述代码中，我们主要是定义了 Request 结构体作为接口入参的基准，而本项目由于并不会太复杂，所以直接放在了 service 层中便于使用，若后续业务不断增长，程序越来越复杂，service 也冗杂了，可以考虑将抽离一层接口校验层，便于解耦逻辑。

另外我们还在 service 进行了一些简单的逻辑封装，在应用分层中，service 层主要是针对业务逻辑的封装，如果有一些业务聚合和处理可以在该层进行编码，同时也能较好的隔离上下两层的逻辑。

### 2.3.2 新建业务错误码

我们在项目的 `pkg/errcode` 下新建 module_code.go 文件，针对标签模块，写入如下错误代码：

```go
var (
	ErrorGetTagListFail = NewError(20010001, "获取标签列表失败")
	ErrorCreateTagFail  = NewError(20010002, "创建标签失败")
	ErrorUpdateTagFail  = NewError(20010003, "更新标签失败")
	ErrorDeleteTagFail  = NewError(20010004, "删除标签失败")
	ErrorCountTagFail   = NewError(20010005, "统计标签失败")
)
```

### 2.3.3 新建路由方法

我们打开 `internal/routers/api/v1` 项目目录下的 tag.go 文件，写入如下代码：

```go
func (t Tag) List(c *gin.Context) {
	param := service.TagListRequest{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, &param)
	if !valid {
		global.Logger.Errorf("app.BindAndValid errs: %v", errs)
		response.ToErrorResponse(errcode.InvalidParams.WithDetails(errs.Errors()...))
		return
	}

	svc := service.New(c.Request.Context())
	pager := app.Pager{Page: app.GetPage(c), PageSize: app.GetPageSize(c)}
	totalRows, err := svc.CountTag(&service.CountTagRequest{Name: param.Name, State: param.State})
	if err != nil {
		global.Logger.Errorf("svc.CountTag err: %v", err)
		response.ToErrorResponse(errcode.ErrorCountTagFail)
		return
	}
	
	tags, err := svc.GetTagList(&param, &pager)
	if err != nil {
		global.Logger.Errorf("svc.GetTagList err: %v", err)
		response.ToErrorResponse(errcode.ErrorGetTagListFail)
		return
	}

	response.ToResponseList(tags, totalRows)
	return
}
```

在上述代码中，我们完成了获取标签列表接口的处理方法，我们在方法中完成了入参校验和绑定、获取标签总数、获取标签列表、 序列化结果集等四大功能板块的逻辑串联和日志、错误处理。

需要注意的是入参校验和绑定的处理代码基本都差不多，因此在后续代码中不再重复，我们继续写入创建标签、更新标签、删除标签的接口处理方法，如下：

```go
func (t Tag) Create(c *gin.Context) {
	param := service.CreateTagRequest{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, &param)
	if !valid {...}

	svc := service.New(c.Request.Context())
	err := svc.CreateTag(&param)
	if err != nil {
		global.Logger.Errorf("svc.CreateTag err: %v", err)
		response.ToErrorResponse(errcode.ErrorCreateTagFail)
		return
	}

	response.ToResponse(gin.H{})
	return
}

func (t Tag) Update(c *gin.Context) {
	param := service.UpdateTagRequest{ID: convert.StrTo(c.Param("id")).MustUInt32()}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, &param)
	if !valid {...}

	svc := service.New(c.Request.Context())
	err := svc.UpdateTag(&param)
	if err != nil {
		global.Logger.Errorf("svc.UpdateTag err: %v", err)
		response.ToErrorResponse(errcode.ErrorUpdateTagFail)
		return
	}

	response.ToResponse(gin.H{})
	return
}

func (t Tag) Delete(c *gin.Context) {
	param := service.DeleteTagRequest{ID: convert.StrTo(c.Param("id")).MustUInt32()}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, &param)
	if !valid {...}

	svc := service.New(c.Request.Context())
	err := svc.DeleteTag(&param)
	if err != nil {
		global.Logger.Errorf("svc.DeleteTag err: %v", err)
		response.ToErrorResponse(errcode.ErrorDeleteTagFail)
		return
	}

	response.ToResponse(gin.H{})
	return
}
```

## 2.3 文件服务

本章节增加**上传文件**的功能

### 2.3.1 新增配置

打开项目下的`configs/config.yaml`配置文件，新增上传相关的配置

```yaml
App:
	...
	UploadSavePath: storage/uploads
	UploadServerUrl: http://127.0.0.1:8000/static
	UploadImageMaxSize: 5
	UploadImageAllowExts: 
		- .jpg
		- .jpeg
		- .png
```

上面这段代码一共有四项，分别代表的作用如下：

- UploadSavePath: 上传文件的最终保存目录
- UploadServerUrl：上传文件后用于展示的文件服务地址
- UploadImageMaxSize：上传文件所允许的最大空间大小（MB）
- UploadImageAllowExts：上传文件所允许的文件后缀