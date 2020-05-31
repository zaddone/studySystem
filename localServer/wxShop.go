package main
import(
	"fmt"
	"io"
	"net/url"
	"github.com/gin-gonic/gin"
	"github.com/zaddone/studySystem/request"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"bytes"
	"strings"
)

var(
	GoodsDB = "Category.db"
	WxCategory = []byte("WxCategory")
)

//func GoodsWithAlibabaToWX(obj interface{}){
//
//	db := obj.(map[string]interface{})
//
//}

func openGoodsDB(write bool,hand func(*bolt.Tx)error)error{
	db,err := bolt.Open(GoodsDB,0600,nil)
	if err != nil {
		return err
	}
	defer db.Close()
	t,err := db.Begin(write)
	if err != nil {
		return err
	}
	if write {
		defer t.Commit()
	}
	return hand(t)
}

func GetToken(hand func(string)error) error {
	u := url.Values{}
	addSign(&u)
	return request.ClientHttp_("https://www.zaddone.com/wxserver/token?"+u.Encode(),"GET",nil,nil,func(body io.Reader,re int)error{
		db,err := ioutil.ReadAll(body)
		if err != nil {
			return err
		}
		//fmt.P
		return hand(string(db))
	})
}

func ShowCategory(name string,hand func(string,interface{})error)error{
	return openGoodsDB(true,func(t *bolt.Tx)error{
		b := t.Bucket(WxCategory)
		if b == nil {
			return nil
		}
		if len(name) > 0 {
			b_ := b.Bucket([]byte(name))
			if b_ == nil {
				return nil
			}
			val := map[string]interface{}{}
			err := b_.ForEach(func(k,v []byte)error{
				val[string(k)] = string(v)
				return nil
			})
			if err != nil {
				return err
			}
			return hand(name,val)
		}
		c := b.Cursor()
		for b_,_ := c.First();b_ != nil;b_,_ = c.Next() {
			val := map[string]interface{}{}
			//bv := b.Bucket(b_)
			b__:= b.Bucket(b_)
			err :=  b__.ForEach(func(k,v []byte)error{
				val[string(k)] = string(v)
				return nil
			})
			if err != nil {
				return err
			}
			//if val["children"] != nil {
			//err = GetCategorySku(val["id"].(string),func(vb interface{})error{
			//	fmt.Println(vb)
			//	sku := vb.(map[string]interface{})["sku_table"]
			//	if sku != nil {
			//		db,err := json.Marshal(sku)
			//		if err != nil {
			//			return err
			//		}
			//		return b__.Put([]byte("sku"),db)
			//	}
			//	return nil
			//})
			//if err != nil {
			//	return err
			//}
			//err = GetCategoryInfo(val["id"].(string),func(vb interface{})error{
			//	fmt.Println(vb)
			//	sku := vb.(map[string]interface{})["properties"]
			//	if sku != nil {
			//		db,err := json.Marshal(sku)
			//		if err != nil {
			//			return err
			//		}
			//		return b__.Put([]byte("properties"),db)
			//	}
			//	return nil
			//})
			//}else{
			err = hand(string(b_),val)
			if err != nil {
				return err
			}
			//}

		}
		return nil
	})

}
func GetCategoryInfo(id string,hand func(interface{})error)error{
	return Request(
		"https://api.weixin.qq.com/merchant/category/getproperty",
		map[string]interface{}{"cate_id":id},
		hand)

}
func GetCategorySku(id string,hand func(interface{})error)error{
	//dbMap := map[string]interface{}{"cate_id":id}
	return Request(
		"https://api.weixin.qq.com/merchant/category/getsku",
		map[string]interface{}{"cate_id":id},
		hand)
}
func Request(uri string,dbMap interface{},hand func(interface{})error) error {
	db_ ,err := json.Marshal(dbMap)
	if err != nil {
		return err
	}
	return GetToken(func(token string)error{
	u := url.Values{}
	u.Set("access_token",token)
	return request.ClientHttp_(uri+"?"+u.Encode(),"POST",bytes.NewReader(db_),nil,
	func(body io.Reader,re int)error{
		var val interface{}
		err := json.NewDecoder(body).Decode(&val)
		if err != nil {
			return err
		}
		return hand(val)
	})
	})

}

func DownCategory(id string,hand func(interface{})error)error{

	u := url.Values{}
	return GetToken(func(token string)error{
		u.Set("access_token",token)
		//fmt.Println(u)
		dbMap := map[string]interface{}{
			"cate_id":id,
		}
		db_ ,err := json.Marshal(dbMap)
		if err != nil {
			return err
		}
		uri := "https://api.weixin.qq.com/merchant/category/getsub?"+u.Encode()
		fmt.Println(uri,string(db_))
		return request.ClientHttp_(
			uri,"POST",bytes.NewReader(db_),nil,
		func(body io.Reader,re int)error{
			var val interface{}
			err := json.NewDecoder(body).Decode(&val)
			if err != nil {
				return err
			}
			return hand(val)
			//return nil
		})
	})

}

func handCategory(pidName string,db interface{},pb,b *bolt.Bucket)error {

	db_ := db.(map[string]interface{})["cate_list"]
	if db_ == nil {
		return nil
	}
	var li []string
	for _,d := range db_.([]interface{}){
		d_ := d.(map[string]interface{})
		//go func(d_ map[string]interface{}){
		//fmt.Println(d_)
		name := d_["name"].(string)
		li = append(li,name)
		id := d_["id"].(string)
		//b.CreateBucketIfNotExists(
		b_,err := b.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			fmt.Println(err)
			return err
		}
		if len(pidName)>0{
			err := b_.Put([]byte("pName"),[]byte(pidName))
			if err != nil {
				return err
				//return
			}
		}
		err = b_.Put([]byte("id"),[]byte(id))
		if err != nil {
			fmt.Println(err)
			return err
		}
		err = DownCategory(id,func(__db interface{})error{
			return handCategory(name,__db,b_,b)
		})
		if err != nil {
			fmt.Println(err)
			return err
			//panic(err)
			//return err
		}
		//hand(d_["name"].(string),d_["id"].(string))
		//}(d.(map[string]interface{}))
	}
	if pb!=nil && len(li)>0{
		pb.Put([]byte("children"),[]byte(strings.Join(li,",")))
	}
	//pb := b.Bucket([]byte(pidName))
	//c.JSON(http.StatusOK,db)
	return nil
}

func init(){
	shop := Router.Group("wxshop")
	shop.GET("category/show",func(c *gin.Context){
		var li []interface{}
		err := ShowCategory(c.Query("key"),func(name string,db interface{})error{
			li = append(li,map[string]interface{}{"name":name,"val":db})
			//https://api.weixin.qq.com/merchant/category/getsku?access_token=ACCESS_TOKEN
			return nil
		})
		if err != nil {
			c.JSON(http.StatusFound,err)
		}
		c.JSON(http.StatusOK,li)
		return
	})
	shop.GET("category/down",func(c *gin.Context){
		err := openGoodsDB(true,func(t *bolt.Tx)error{
			rootCat:="1"
			b,err := t.CreateBucketIfNotExists(WxCategory)
			if err != nil {
				return err
			}
			return DownCategory(rootCat,func(db interface{})error{
				return handCategory("",db,nil,b)
			})
		})
		if err != nil {
			c.JSON(http.StatusFound,err)
		}
		c.String(http.StatusOK,"success")
	})
}
