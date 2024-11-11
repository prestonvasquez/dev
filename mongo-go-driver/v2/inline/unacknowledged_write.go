package main

//func main() {
//	poolMonitor := &event.PoolMonitor{
//		Event: func(pe *event.PoolEvent) {
//			if pe.Type == event.ConnectionCheckedIn {
//			}
//		},
//	}
//
//	client, err := mongo.Connect(options.Client().SetMonitor(newMonitor()))
//	if err != nil {
//		panic(err)
//	}
//
//	defer func() { _ = client.Disconnect(context.Background()) }()
//
//	coll := client.Database("db").Collection("coll", options.Collection().SetWriteConcern(writeconcern.Unacknowledged()))
//
//	_, err = coll.InsertOne(context.Background(), bson.D{{"x", 1}})
//	if err != nil {
//		panic(err)
//	}
//}
