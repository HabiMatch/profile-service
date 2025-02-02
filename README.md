# HabiMatch
HabiMatch is an tech solution which provides ease to user who wants to find a roomate or a room. 

## Profile-Service (GoLang)
Profile-Service handles the all details related to profiles including Profile Images, Keeper-(Who has a room and wants a roommate), Seeker-(is a student or other individual wants a room) and other details.

### .env
```
PORT=8080
DB_HOST=your_db-your_service.g.aivencloud.com  // I prefered Aiven Cloud
DB_PORT=28096
DB_USER=avnadmin
DB_PASSWORD=your_db_password
DB_NAME=defaultdb
S3_REGION=ap-south-1
S3_BUCKET_NAME=habimatch-profile
S3_PROFILE_FOLDER_NAME=profile_pictures
S3_ROOM_FOLDER_NAME=room_images
ROOM_IMAGES_COUNT=3
AWS_ACCESS_KEY=your_Access_key
AWS_SECRET_KEY=your_Secret_key
JWT_SECRET=your_jwt_secret
PROFILE_PICTURE_URL=https://habimatch-profile.s3.ap-south-1.amazonaws.com/profile_pictures/
```

_**You can use any db viewer to see the data. I used pgAdmin4**_.

We use formdata to send data to server. userinfo contains json.

### Operations

```
operation :
create_profile
update_profile
update_profile_picture
update_geolocation
keeper_profile
update_keeper_profile
delete_keeper_profile
seeker_profile
update_seeker_profile
delete_seeker_profile
```
How to send data to server **create_profile, keeper_profile, seeker_profile** relative to their schema :

```
//form-data
operation : create_profile
profile_picture : Your_Profile_Image
userinfo : {
			UserID:         "user_test",
			FirstName:      "John",
			LastName:       "Doe",
			Gender:         "Male",
			Occupation:     "Software Engineer",
			Address:        "123 Elm Street, Springfield, USA",
			Contactno:      "+1234567890",
			Description:    "Tech enthusiast who loves coding and exploring new places.",
			Latitude:       37.7749,
			Longitude:      -122.4194,
			Selftags:       pq.StringArray{"Non-Smoker", "Early Bird", "Music Lover"},
		}
```
## Our Vision
Our vision is to create an android application using Flutter so that we can create these profiles using that application. Which aims to create a scaleable profile service.
