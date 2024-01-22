package git

import (
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/google/go-github/v43/github"
	"github.com/Traced/a0/utils"
)

// CreateRepo 创建一个新的仓库
func CreateRepo(client *Client, name, desc string, private bool) *github.Repository {
	repo, _, err := client.Client.Repositories.Create(
		client.ctx,
		"",
		&github.Repository{
			Name:        &name,
			Private:     &private,
			Description: &desc,
			AutoInit:    github.Bool(true),
		},
	)
	if err != nil {
		log.Println("create new repo err:", err)
		return nil
	}
	return repo
}

// DeleteRepo 删除一个仓库
func DeleteRepo(client *Client, name, owner string) *github.Response {
	r, err := client.Client.Repositories.Delete(client.ctx, owner, name)
	if err != nil {
		log.Println("delete repo err:", err)
		return nil
	}
	return r
}

// 创建一个新文件
func createFile(client *Client, owner, repo, branch, message, filepath string, content []byte) *github.RepositoryContentResponse {
	r, _, err := client.Client.Repositories.CreateFile(
		client.ctx, owner, repo, filepath,
		&github.RepositoryContentFileOptions{
			Content: content,
			Message: &message,
			Branch:  &branch,
		})
	if err != nil {
		log.Printf("create file %s of %s err: %s", filepath, repo, err)
		return nil
	}
	return r
}

// 更新一个文件
func updateFile(client *Client, owner, repo, branch, message, filepath, sha string, content []byte) *github.RepositoryContentResponse {
	r, _, err := client.Client.Repositories.UpdateFile(
		client.ctx, owner, repo, filepath,
		&github.RepositoryContentFileOptions{
			Branch:  &branch,
			Content: content,
			Message: &message,
			SHA:     &sha,
		})
	if err != nil {
		log.Printf("update file %s of %s err: %s", filepath, repo, err)
		return nil
	}
	return r
}

var (
	// UpdateFile 更新一个文件
	// 如果文件不存在就会调用 createFile 函数来创建
	UpdateFile = CreateORUpdateFile
	// CreateFile 创建一个文件
	// 如果文件存在就会调用 updateFile 函数来创建
	CreateFile = CreateORUpdateFile
)

// CreateORUpdateFile 创建一个新文件到指定仓库, 如果
func CreateORUpdateFile(client *Client, owner, repo, branch, message, filepath string, content []byte) *github.RepositoryContentResponse {
	file := GetFileContent(client, owner, repo, branch, filepath)
	// 文件未找到，直接创建一个新文件
	if file == nil {
		return createFile(client, owner, repo, branch, message, filepath, content)
	}
	// 文件存在，直接更新
	return updateFile(client, owner, repo, branch, message, filepath, *file.SHA, content)
}

// DeleteFile 从指定分支删除一个文件
func DeleteFile(client *Client, owner, repo, branch, message, filepath string) *github.RepositoryContentResponse {
	file := GetFileContent(client, owner, repo, branch, filepath)
	// 文件未找到
	if file == nil {
		log.Println("file not found in branch", branch)
		return nil
	}
	r, _, err := client.Client.Repositories.DeleteFile(
		client.ctx, owner, repo, filepath,
		&github.RepositoryContentFileOptions{
			Branch:  &branch,
			Message: &message,
			SHA:     file.SHA,
		})
	if err != nil {
		log.Printf("update file %s of %s err: %s", filepath, repo, err)
		return nil
	}
	return r
}

// 获取指定文件路径信息
func getContent(client *Client, owner, repo, branch, filepath string) (fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, response *github.Response, err error) {
	fileContent, directoryContent, response, err = client.Client.Repositories.GetContents(
		client.ctx, owner, repo, filepath,
		&github.RepositoryContentGetOptions{
			Ref: branch,
		})
	return
}

// GetFileContent 获取文件信息
// 不存在会返回一个 nil
func GetFileContent(client *Client, owner, repo, branch, filepath string) *github.RepositoryContent {
	file, _, _, err := getContent(client, owner, repo, branch, filepath)
	if err != nil {
		log.Println("get repo file content err:", err)
		return nil
	}
	return file
}

// HasFile 判断文件是否存在
func HasFile(client *Client, owner, repo, branch, filepath string) bool {
	return GetFileContent(client, owner, repo, branch, filepath) != nil
}

func FetchGitTrees(client *Client, dirs []*github.RepositoryContent, owner, repo, branch, repoBasePath string) {
	l := len(dirs)
	// 目录里没东西
	if 1 > l {
		return
	}
	// 创建对应数量协程去拉取
	var (
		gc = client.Client
		wg = new(sync.WaitGroup)
		// 目录分隔符
		sep = string(os.PathSeparator)
	)
	wg.Add(l)
	for _, item := range dirs {
		go func(item *github.RepositoryContent) {
			// {base path} / repo name /  item.path
			storePath := repoBasePath + sep + repo + sep + *item.Path
			log.Println("fetching:", *item.Path)
			// 创建文件或者文件夹
			switch *item.Type {
			case "file":
				// 获取 blob 数据，这里包含文件内容
				b, _, err := gc.Git.GetBlob(client.ctx, owner, repo, *item.SHA)
				if err != nil {
					log.Println("fetch blob err:", *item.Name, err)
					break
				}
				createPullFile(storePath, b.GetContent())
			case "dir":
				createPullDir(storePath)
				// 递归拉取目录
				_, dirs, _, err := getContent(client, owner, repo, branch, *item.Path)
				if err != nil {
					log.Println("fetch deep dir err:", *item, err)
					break
				}
				FetchGitTrees(client, dirs, owner, repo, branch, repoBasePath)
			}
			wg.Done()
		}(item)
	}
	wg.Wait()
}

func Pull(client *Client, owner, repo, branch, basePath string) error {
	_, dirs, _, err := getContent(client, owner, repo, branch, "")
	if err != nil {
		log.Println("pull err:", err)
		return err
	}
	FetchGitTrees(client, dirs, owner, repo, branch, basePath)
	log.Println("pulled.")
	return nil
}

func createPullDir(path string) {
	if utils.IsDir(path) {
		log.Println("create dir err:", path, "already exists !")
		return
	}
	utils.Mkdir(path)
}

func createPullFile(filepath string, base64 string) bool {
	var (
		data = utils.Base64DecodeToByte(base64)
		err  = ioutil.WriteFile(filepath, data, 0644)
	)
	if err != nil {
		log.Println("create file err:", filepath, err)
		return false
	}
	return true
}
