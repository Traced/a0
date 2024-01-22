package git

import (
	"github.com/google/go-github/v43/github"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

var (
	DefaultBranch = "main"
)

// NewClient 创建一个新的 GitHub 客户端实例
func NewClient(user, token, repo string) *Client {
	var (
		ctx = context.Background()
		tc  = oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		))
	)
	return &Client{
		ctx:    ctx,
		token:  token,
		owner:  user,
		repo:   repo,
		branch: DefaultBranch,
		Client: github.NewClient(tc),
	}
}

type Client struct {
	token, owner, repo, branch, path string
	ctx                              context.Context
	Client                           *github.Client
}

// SetPath 设置项目存储路径
func (c *Client) SetPath(path string) *Client {
	c.path = path
	return c
}

// GetPath 获取当前设置的项目路径
func (c *Client) GetPath() string {
	return c.path
}

// SetRepo 设置当前使用的仓库
func (c *Client) SetRepo(repo string) *Client {
	c.repo = repo
	return c
}

// GetRepo 获取当前使用的仓库
func (c *Client) GetRepo() string {
	return c.repo
}

// SetBranch 选择一个分支
func (c *Client) SetBranch(branch string) *Client {
	c.branch = branch
	return c
}

// GetBranch 获取当前分支
func (c *Client) GetBranch() string {
	return c.branch
}

// SetOwner 设置当前仓库的拥有者
func (c *Client) SetOwner(owner string) *Client {
	c.owner = owner
	return c
}

// GetOwner 获取当前使用的仓库
func (c *Client) GetOwner() string {
	return c.owner
}

func (c *Client) Pull() error {
	return Pull(c, c.owner, c.repo, c.branch, c.path)
}

func (c *Client) Push() {

}

// NewRepo 新建一个仓库
func (c *Client) NewRepo(name, desc string, private bool) *github.Repository {
	return CreateRepo(c, name, desc, private)
}

// DeleteRepo 删除一个仓库
func (c *Client) DeleteRepo(name string) *github.Response {
	return DeleteRepo(c, name, c.owner)
}

// NewFile 创建新文件
func (c *Client) NewFile(message, filepath string, content []byte) *github.RepositoryContentResponse {
	return CreateFile(c, c.owner, c.repo, c.branch, message, filepath, content)
}

// DeleteFile 删除文件
func (c *Client) DeleteFile(message, filepath string) *github.RepositoryContentResponse {
	return DeleteFile(c, c.owner, c.repo, c.branch, message, filepath)
}

// UpdateFile 更新文件
func (c *Client) UpdateFile(message, filepath string, content []byte) *github.RepositoryContentResponse {
	return UpdateFile(c, c.owner, c.repo, c.branch, message, filepath, content)
}

// GetFileContent 获取文件内容
func (c *Client) GetFileContent(filepath string) *github.RepositoryContent {
	return GetFileContent(c, c.owner, c.repo, c.branch, filepath)
}
