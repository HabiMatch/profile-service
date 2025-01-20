package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"mime/multipart"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func UploadToS3(file multipart.File, folderName string, fileName string) (string, error) {

	fileName = strings.ReplaceAll(fileName, " ", "")
	s3Region := os.Getenv("S3_REGION")
	s3Bucket := os.Getenv("S3_BUCKET_NAME")

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s3Region),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY"),
			os.Getenv("AWS_SECRET_KEY"),
			"",
		),
	})
	if err != nil {
		return "", err
	}

	svc := s3.New(sess)
	buf := bytes.NewBuffer(nil)
	if _, err := buf.ReadFrom(file); err != nil {
		return "", err
	}

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(folderName + "/" + fileName),
		Body:   bytes.NewReader(buf.Bytes()),
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s/%s", s3Bucket, s3Region, folderName, fileName)
	return url, nil
}
func DeleteFromS3(fileURL string) error {
	s3Region := os.Getenv("S3_REGION")
	s3Bucket := os.Getenv("S3_BUCKET_NAME")

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s3Region),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY"),
			os.Getenv("AWS_SECRET_KEY"),
			"",
		),
	})
	if err != nil {
		return err
	}

	svc := s3.New(sess)

	// Extract the file key from the URL
	fileKey := fileURL[len(fmt.Sprintf("https://%s.s3.%s.amazonaws.com/", s3Bucket, s3Region)):]

	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return err
	}

	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return err
	}

	return nil
}

func IsImageFile(file multipart.File) (bool, error) {
	buf := bytes.NewBuffer(nil)
	if _, err := buf.ReadFrom(file); err != nil {
		return false, fmt.Errorf("failed to read file: %v", err)
	}

	_, _, err := image.Decode(buf)
	if err != nil {
		return false, nil
	}

	return true, nil
}

func ConvertToJPEG(file multipart.File, userID string) (*os.File, string, error) {
	// Decode the image from the file
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %v", err)
	}

	// Create a temporary file to store the JPEG
	tempFileName := fmt.Sprintf("%s.jpeg", userID)
	tempFile, err := os.CreateTemp("", "converted_*.jpeg")
	if err != nil {
		return nil, "", fmt.Errorf("unable to create temporary file: %v", err)
	}

	// Encode the image to JPEG format
	err = jpeg.Encode(tempFile, img, &jpeg.Options{Quality: 85})
	if err != nil {
		return nil, "", fmt.Errorf("failed to encode JPEG: %v", err)
	}

	// Close and reopen the temp file for reading
	if err := tempFile.Close(); err != nil {
		return nil, "", err
	}
	tempFile, err = os.Open(tempFile.Name())
	if err != nil {
		return nil, "", err
	}

	return tempFile, tempFileName, nil
}
