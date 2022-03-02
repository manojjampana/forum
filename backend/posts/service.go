package posts

import (
	"context"
	"fmt"
	"net/http"
	"simple-reddit/common"
	"simple-reddit/configs"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const POST_ROUTE_PREFIX = "/post"

const PostsCollectionName string = "posts"

var postCollection *mongo.Collection = configs.GetCollection(configs.MongoClient, PostsCollectionName)
var validate = validator.New()

func CreatePost() gin.HandlerFunc {
	return func(c *gin.Context) {
		var post CreatePostRequest
		if err := c.BindJSON(&post); err != nil {
			c.JSON(
				http.StatusBadRequest,
				common.APIResponse{
					Status:  http.StatusBadRequest,
					Message: common.API_FAILURE,
					Data:    map[string]interface{}{"data": err.Error()}},
			)
			return
		}
		if validationErr := validate.Struct(&post); validationErr != nil {
			c.JSON(
				http.StatusBadRequest,
				common.APIResponse{
					Status:  http.StatusBadRequest,
					Message: common.API_FAILURE,
					Data:    map[string]interface{}{"data": validationErr.Error()}},
			)
			return
		}
		result, err := createPostInDB(post)
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				common.APIResponse{
					Status:  http.StatusInternalServerError,
					Message: common.API_ERROR,
					Data:    map[string]interface{}{"data": err.Error()}},
			)
			return
		}

		c.JSON(
			http.StatusCreated,
			common.APIResponse{
				Status:  http.StatusCreated,
				Message: common.API_SUCCESS,
				Data:    map[string]interface{}{"data": result}},
		)

	}
}

func GetPosts() gin.HandlerFunc {
	return func(c *gin.Context) {
		var postReq GetPostRequest

		// validate the request body
		if err := c.BindJSON(&postReq); err != nil {
			c.JSON(
				http.StatusBadRequest,
				configs.APIResponse{
					Status:  http.StatusBadRequest,
					Message: configs.API_FAILURE,
					Data:    map[string]interface{}{"data": err.Error()}},
			)
			return
		}
		fmt.Println("bind passed")
		// use the validator library to validate required fields
		if validationErr := validate.Struct(&postReq); validationErr != nil {
			c.JSON(
				http.StatusBadRequest,
				configs.APIResponse{
					Status:  http.StatusBadRequest,
					Message: configs.API_FAILURE,
					Data:    map[string]interface{}{"data": validationErr.Error()}},
			)
			return
		}
		fmt.Println("validation done")
		postDetails, err := retrievePostDetails(postReq)
		fmt.Println(postDetails)
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				configs.APIResponse{
					Status:  http.StatusInternalServerError,
					Message: configs.API_ERROR,
					Data:    map[string]interface{}{"data": err.Error()}},
			)
			return
		}

		if len(postDetails) > 0 {
			c.JSON(
				http.StatusOK,
				configs.APIResponse{
					Status:  http.StatusOK,
					Message: configs.API_SUCCESS,
					Data:    map[string]interface{}{"posts": postDetails}},
			)
			return
		} else {
			c.JSON(
				http.StatusOK,
				configs.APIResponse{
					Status:  http.StatusNotFound,
					Message: configs.API_SUCCESS,
					Data:    map[string]interface{}{"posts": postDetails}},
			)
			return
		}
	}
}

func createPostInDB(post CreatePostRequest) (result *mongo.InsertOneResult, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	newPost := ConvertPostRequestToPostDBModel(post)
	if err != nil {
		return result, err
	}
	result, err = postCollection.InsertOne(ctx, newPost)
	return result, err
}

func retrievePostDetails(postReq GetPostRequest) ([]PostResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var posts []PostDBModel
	var postResp []PostResponse
	filter := bson.D{primitive.E{Key: "community_id", Value: postReq.CommunityID}, primitive.E{Key: "username", Value: postReq.UserName}}
	cursor, err := postCollection.Find(ctx, filter)
	if err = cursor.All(ctx, &posts); err != nil {
		return postResp, err
	}
	fmt.Println("posts")
	fmt.Println(posts)
	for _, post := range posts {
		item, err := ConvertPostDBModelToPostResponse(post)
		if err != nil {
			return postResp, err
		}
		fmt.Print(item)
		postResp = append(postResp, item)

	}
	return postResp, err
}

func Routes(router *gin.Engine) {
	router.POST(POST_ROUTE_PREFIX+"/create", CreatePost())
	router.GET(POST_ROUTE_PREFIX+"/get", GetPosts())

}