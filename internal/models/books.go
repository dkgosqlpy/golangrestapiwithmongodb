package models

type Book struct {
	Name      string `bson:"name" form:"name" binding:"required,min=3"`
	Author    string `bson:"author" form:"author" binding:"required,min=3"`
	PageCount int    `bson:"page_count" form:"count" binding:"required,min=1"`
}

type Author struct {
	FullName string `bson:"full_name"`
}

type AuthorBooks struct {
	FullName string `bson:"full_name"`
	Books    []Book
}
