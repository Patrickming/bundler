package config

import (
	"github.com/joho/godotenv"
)

// LoadEnv 加载 .env 文件中的环境变量
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}
}

// // GetMongoClient 创建并返回一个 MongoDB 客户端
// func GetMongoClient() (*mongo.Client, error) {
// 	mongoURI := os.Getenv("MONGO_URI")

// 	clientOptions := options.Client().ApplyURI(mongoURI)
// 	client, err := mongo.Connect(context.TODO(), clientOptions)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = client.Ping(context.TODO(), nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return client, nil
// }
