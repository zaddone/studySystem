package dictionary
import(
	"encoding/binary"
	"github.com/boltdb/bolt"
)

var (
	db *bolt.DB
	fileName string = "dict.db"
)

func open(h func(db *bolt.DB)){

	db,err := bolt.Open(fileName,0600,nil)
	if err != nil {
		panic(err)
	}
	h(db)
	db.Close()

}
//func split (sentence string)

func update(sentence string){
	open(func(db *bolt.DB){
		db.Update(func(tx *bolt.Tx) error{
			l := len(dic)
			if l>255 {
				l = 255
			}
			b,err := tx.CreateBucketIfNotExists([]byte{byte(l)})
			if err != nil {
				return err
			}
			b.Put([]byte(dic),[]byte{0})
		})
	})
}
