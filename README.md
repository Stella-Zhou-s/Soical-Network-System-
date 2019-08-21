Social Network System 
- Built a scalable geo-based social network in Go to handle posts and deployed to Google Cloud (GKE) for better scaling
- Utilized ElasticSearch (GCE) to provide geo-location-based search functions such that users can search nearby posts within a distance (e.g. 200km)
- Used Google Dataflow to implement a daily dump of posts to BigQuery table for offline analysis
- Implemented basic token-based registration/login/logout flow with React Router v4 and server-side user authentication with JWT


# Overview of project
- Web services in **Golang** to handle posts, seardh and user login, logout are deployed to **Google App Engine(GAE flex)**.
- **ElasticSearch** in **GCE** provide storage and geo-location based search for user nearby posts within a distance.
- Use **Google Dataflow** to dump posts from **BigTable** to **BigQuery** for offline analysis
- Use **Google Cloud Storage(GCS)** to store post image.
- Use **OAuth** 2.0 to support token based authentication.
- Use **Redis(lazy-loading)** to improve read performance with a little data consistency sacrifice.
![image](https://user-images.githubusercontent.com/38120488/38523155-a033d86a-3c18-11e8-8912-706ab4ec3528.png)

# Services provided and API design
- **/signup**
  * save to elasticSearch.
- **/login**
  * check login credential in elasticSearch, if correct return token
- **/search** - search nearby posts.
  1. have token-based authentication first.
  2. search in redis cache, if not found then search elasticSearch(lazy-loading).
  3. use  `"type" : "geo_point"` to map (lat, lon) to geo_point, ES will use geo-indexing to search(**KD tree**) nearby posts.
- **/post**
  1. save post image in GCS. 
  2. save post info in ElasticSearch, bigTable(optional).

![image](https://user-images.githubusercontent.com/38120488/38536128-6afd7056-3c55-11e8-876e-5fa628a0123b.png)

# Storage
- ElasticSearch(save user and post infos)
  * user info example
  ```json
   {
      "_index" : "around",
      "_type" : "user",
      "_id" : "jack",
      "_score" : 1.0,
      "_source" : {
        "username" : "jack",
        "password" : "jack"
      }
    }
  ```
  * post info example
  ```
  {
      "_index" : "around",
      "_type" : "post",
      "_id" : "b2c32515-c07d-4154-b2b1-6c7ab5e06d42",
      "_score" : 1.0,
      "_source" : {
        "user" : "jack",
        "message" : "Nice star!",
        "location" : {
          "lat" : 44.70415541365263,
          "lon" : -78.12385288120937
        },
        "url" : "https://www.googleapis.com/download/storage/v1/b/.../6c7ab5e06d42?generation=1522902512391320&alt=media"
      }
    }
  ```
  
- Google Cloud Storage
  * elasticserach saves image url, GCS store real image file.
- Redis
  * redis can be simply regarded as key-value store
    * **key**: **lat:lon:range**, range is redius based on (lat,lon) as circle center.
    * **value**: **post info**
- BigTable, BigQuery
  * we can save posts data to BigTable, use Dataflow to pass posts from BigTable to BigQuery for data analysis. See DataFlow code [here](./dataflow/src/main/java/com/around/PostDumpFlow.java).
  * Several cases can be done in BigQuery:
    * get number of posts per user id -- base to detect spam user.
    * find all messages in LA(lat range [33, 34], lon range [-118, -117]) -- base for geo-based filter.
    * find all message with spam words
    
# Implimentation Details
- routing and auth
  * Here use `gorilla/mux` for routing and `dgrijalva/jwt-go` for JWT(JSON Web Token) token based authentication, useful [doc](https://auth0.com/blog/authentication-in-golang/).
- redis cache
  * adopt lazy-loading(load DB after a cache miss) pattern.
  * 30MB RAM, 30 connections for [free](https://redislabs.com/blog/redis-cloud-30mb-ram-30-connections-for-free/).
  * [Sample code](https://github.com/go-redis/redis)
 
 # References
 - [QuickStart](https://cloud.google.com/appengine/docs/flexible/go/quickstart) for Go in GAE flex.
 - [Example](https://github.com/olivere/elastic) elastic search in go.
 - [Example](https://github.com/GoogleCloudPlatform/golang-samples/blob/master/storage/objects/main.go) writing object to GCS.
 - [Example](https://cloud.google.com/storage/docs/reference/libraries#client-libraries-install-go) client connection to GCS. 
 - [Example](https://cloud.google.com/dataflow/model/bigquery-io#writing-to-bigquery) writing to BigQuery
 - [Example](https://github.com/golang-samples/http/blob/master/fileupload/main.go) Go parseMultipart form
 - [uuid](https://github.com/pborman/uuid) : each post is unique
 - [encoding](https://golang.org/pkg/encoding/json/) json

