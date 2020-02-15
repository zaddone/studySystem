package main
import(
	//"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	//"net/http"
	//"fmt"
	//"time"
	//"bytes"
)
func init(){
	Router := gin.Default()
	Router.Static("/","./")

	go Router.Run(":8001")
}
