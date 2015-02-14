# API document

## Upload APK

### Usage

``` sh
$ curl http://your-domain.com/api/upload_bundle \
    -F token=your-project-api-token \
    -F description='for alpha-test' \
    -F file=@/path/to/your/apk-file
```

### Parameters

|Name|Description|
|:---:|:---:|
|token|**Required.** The API token of your project. You can check it in your project page.|
|description|The description of the APK file.|
|file|**Required.** The path to the APK file.|

### Response

```
{
  "status": 200,
  "message": [
    "Bundle is created!"
  ],
  "content": {
    "file_id": "the ID of APK file on Google Drive",
    "revision": 1,
    "version": "1.0",
    "install_url": "the URL to install the APK file uploaded",
    "qr_code_url": "the URL of the QR code to install the APK file uploaded"
  }
}
```

## Listing Bundle

### Usage

``` sh
$ curl -XGET http://your-domain.com/api/list_bundle \
    -F token=your-project-api-token \
    -F limit=limit_num \
    -F offset=offset-num
```

### Parameters

|Name|Description|
|:---:|:---:|
|token|**Required.** The API token of your project. You can check it in your project page.|
|limit|Specific number for limit.|
|offset|Specific number for offset.|

### Response

```
{
  "status": 200,
  "message": [
    "Bundle List"
  ],
  "content": {
    "total_count": 2,
    "limit": 2,
    "offset": 1,
    "bundles": [
      {
        "file_id": "the ID of APK file on Google Drive",
        "revision": 1,
        "version": "1.0",
        "qr_code_url": "the URL of the QR code to install the APK file uploaded",
        "install_url": "the URL to install the APK file uploaded"
      },
      {
      .
      .
      .
    ]
  }
}
```
