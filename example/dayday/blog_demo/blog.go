package blog_demo

type Category struct {
	ID   int
	Name string
	Slug string
}

type Post struct {
	ID          int32
	Categories  []Category
	Title       string
	Description string
	Slug        string
}
