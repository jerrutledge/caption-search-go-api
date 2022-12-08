package episode

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Episode struct {
	Full_text string `bson:"full_text"`
	Title     string `bson:"title"`
	Yt_id     string `bson:"yt_id"`
}

type SearchResults struct {
	Err     bool      `bson:"error"`
	Results []Episode `bson:"results"`
}

// CREATE
func Create(collection *mongo.Collection, episode Episode) {

	insertResult, err := collection.InsertOne(context.TODO(), episode)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted a single document: ", insertResult.InsertedID)
}

// READ
func Read(collection *mongo.Collection, filter bson.D) Episode {
	var result Episode

	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found a single document: %+v\n", result)
	return result
}

// UPDATE
func Update(collection *mongo.Collection, filter bson.D) {

	update := bson.D{
		{"$set", bson.D{
			{"full_text", "Cry all your tears, your update is here"},
		}},
	}

	updateResult, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
}

// DELETE
func Delete_all(collection *mongo.Collection) {
	deleteResult, err := collection.DeleteMany(context.TODO(), bson.D{{}})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deleted %v documents in the collection\n", deleteResult.DeletedCount)
}

// SEARCH
func Search(collection *mongo.Collection, searchterm string) (error, []Episode) {
	searchStage := bson.D{{"$search", bson.D{{"text", bson.D{{"path", "full_text"}, {"query", searchterm}}}}}}
	projectStage := bson.D{{"$project", bson.D{{"yt_id", 1}, {"full_text", 1}, {"title", 1}, {"_id", 0}}}}
	// specify the amount of time the operation can run on the server
	opts := options.Aggregate().SetMaxTime(5 * time.Second)
	// run pipeline
	cursor, err := collection.Aggregate(context.Background(), mongo.Pipeline{searchStage, projectStage}, opts)
	if err != nil {
		return err, nil
	}
	// print results
	var search_results []Episode
	words := strings.Fields(searchterm)
	for cursor.Next(context.TODO()) {
		var result Episode
		if err := cursor.Decode(&result); err != nil {
			fmt.Println(err)
			// TODO: actually handle error
			continue
		}
		// translate the full text to partial text
		numChars := 130
		for _, w := range words {
			var index int = strings.Index(strings.ToLower(result.Full_text), strings.ToLower(w))
			if index != -1 {
				if len(result.Full_text)-index > numChars {
					result.Full_text = "..." + result.Full_text[index:index+numChars] + "..."
				} else {
					result.Full_text = "..." + result.Full_text[index:]
				}
				break
			}
		}
		// in case no partial text was found include the first numChars chars
		if len(result.Full_text) > numChars+5 {
			result.Full_text = result.Full_text[:numChars] + "..."
		}
		search_results = append(search_results, result)
	}
	return nil, search_results
}
