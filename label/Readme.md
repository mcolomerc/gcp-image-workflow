# Label Serivce

This service uses Cloud Vision API to classify an input image. 

* Source: Cloud Storage image path 

* Output: Creates a copy of the input image building a classification path on the target bucket. It only takes into account the first annotation to create the directory. 
Example, if the Cloud Vision API returns Food as the first annotation, the image will be stored as gs://<output_bucket>/Food/<file_name>.jpg

Request Body: 

* Bucket: Source bucket

* Object: Object path 

* Output: Target bucket

# BUILD AND DEPLOY

* Google Cloud Project configuration

```
export PROJECT_NUMBER="$(gcloud projects describe $(gcloud config get-value project) --format='value(projectNumber)')" 
```

```
export PROJECT=[YOUR-PROJECT-ID]
```

``` 
export SERVICE="label-service" 
```

``` 
export REGION="[YOUR_REGION]"  
```

``` 
export IMAGE_URL="gcr.io/$(gcloud config get-value project)/label-service" 
``` 

* IAM

``` 
gcloud projects add-iam-policy-binding $(gcloud config get-value project) \
    --member="serviceAccount:service-${PROJECT_NUMBER}@gcp-sa-pubsub.iam.gserviceaccount.com"\
    --role='roles/iam.serviceAccountTokenCreator' 
```

* Enabling Cloud Build

```
gcloud services enable cloudbuild.googleapis.com
```

* Enabling Cloud Run
  
```
gcloud services enable run.googleapis.com
```     

* Cloud Run configuration
  
```
gcloud config set run/platform managed
```

```
gcloud config set run/region ${REGION}
```

* BUILD

```
gcloud builds submit --tag ${IMAGE_URL} 
```  

* DEPLOY

``` 
gcloud beta run deploy ${SERVICE} --image ${IMAGE_URL} --platform managed --no-allow-unauthenticated 
```

## Next

- [ ] Handle multiple annotations > /annotation1/annotation2/... 

 
 
