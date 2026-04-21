---
title: 默认模块
language_tabs:
  - shell: Shell
  - http: HTTP
  - javascript: JavaScript
  - ruby: Ruby
  - python: Python
  - php: PHP
  - java: Java
  - go: Go
toc_footers: []
includes: []
search: true
code_clipboard: true
highlight_theme: darkula
headingLevel: 2
generator: "@tarslib/widdershins v4.0.30"

---

# 默认模块

Base URLs:

# Authentication

# interaction

<a id="opIdinteractionDeleteCommentHandler"></a>

## POST DeleteCommentHandler

POST /comment/delete

> Body 请求参数

```yaml
comment_id: ""

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|
|» comment_id|body|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "base_response": {
    "status_code": 0,
    "status_msg": "string"
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» base_response|object|false|none||none|
|»» status_code|integer|true|none||none|
|»» status_msg|string|true|none||none|

<a id="opIdinteractionGetCommentListHandler"></a>

## GET GetCommentListHandler

GET /comment/list

> Body 请求参数

```yaml
video_id: v2046601080455827456
page_number: "1"
page_size: "1"

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|
|» video_id|body|string| 否 |none|
|» page_number|body|string| 否 |none|
|» page_size|body|string| 否 |none|

> 返回示例

> 200 Response

```json
{
  "base_response": {
    "status_code": 0,
    "status_msg": "string"
  },
  "comment_count": 0,
  "comment_list": [
    {
      "Comment_id": "string",
      "content": "string",
      "created_at": "string",
      "deleted_at": "string",
      "like_count": 0,
      "updated_at": "string",
      "user_id": "string",
      "video_id": "string"
    }
  ]
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» base_response|object|false|none||none|
|»» status_code|integer|true|none||none|
|»» status_msg|string|true|none||none|
|» comment_count|integer|false|none||none|
|» comment_list|[object]|false|none||none|
|»» Comment_id|string|true|none||none|
|»» content|string|true|none||none|
|»» created_at|string|true|none||none|
|»» deleted_at|string|true|none||none|
|»» like_count|integer|true|none||none|
|»» updated_at|string|true|none||none|
|»» user_id|string|true|none||none|
|»» video_id|string|true|none||none|

<a id="opIdinteractionCommentVideoHandler"></a>

## POST CommentVideoHandler

POST /comment/publish

> Body 请求参数

```yaml
video_id: v2046601080455827456
comment_text: hello

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|
|» video_id|body|string| 是 |none|
|» comment_text|body|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "base_response": {
    "status_code": 0,
    "status_msg": "string"
  },
  "comment_id": "string"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» base_response|object|false|none||none|
|»» status_code|integer|true|none||none|
|»» status_msg|string|true|none||none|
|» comment_id|string|false|none||none|

<a id="opIdinteractionLikeVideoHandler"></a>

## POST LikeVideoHandler

POST /like/action

> Body 请求参数

```yaml
video_id: v2046601080455827456
action_type: 0

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|
|» video_id|body|string| 是 |none|
|» action_type|body|integer| 是 |none|

> 返回示例

> 200 Response

```json
{
  "base_response": {
    "status_code": 0,
    "status_msg": "string"
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» base_response|object|false|none||none|
|»» status_code|integer|true|none||none|
|»» status_msg|string|true|none||none|

<a id="opIdinteractionGetLikeListHandler"></a>

## GET GetLikeListHandler

GET /like/list

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|

> 返回示例

> 200 Response

```json
{
  "base_response": {
    "status_code": 0,
    "status_msg": "string"
  },
  "like_count": 0,
  "video_list": [
    {
      "author_id": "string",
      "cover_url": "string",
      "created_at": "string",
      "deleted_at": "string",
      "description": "string",
      "title": "string",
      "updated_at": "string",
      "video_id": "string",
      "video_url": "string"
    }
  ]
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» base_response|object|false|none||none|
|»» status_code|integer|true|none||none|
|»» status_msg|string|true|none||none|
|» like_count|integer|false|none||none|
|» video_list|[object]|false|none||none|
|»» author_id|string|true|none||none|
|»» cover_url|string|true|none||none|
|»» created_at|string|true|none||none|
|»» deleted_at|string|true|none||none|
|»» description|string|true|none||none|
|»» title|string|true|none||none|
|»» updated_at|string|true|none||none|
|»» video_id|string|true|none||none|
|»» video_url|string|true|none||none|

# communication

<a id="opIdcommunicationGetFansListHandler"></a>

## GET GetFansListHandler

GET /follower/list

> Body 请求参数

```yaml
page_number: "1"
page_size: "2"

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|
|» page_number|body|string| 否 |none|
|» page_size|body|string| 否 |none|

> 返回示例

> 200 Response

```json
{
  "base_response": {
    "status_code": 0,
    "status_msg": "string"
  },
  "fans_count": 0,
  "fans_list": [
    {
      "created_at": "string",
      "deleted_at": "string",
      "password": "string",
      "photo_url": "string",
      "updated_at": "string",
      "user_id": "string",
      "username": "string"
    }
  ]
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» base_response|object|false|none||none|
|»» status_code|integer|true|none||none|
|»» status_msg|string|true|none||none|
|» fans_count|integer|false|none||none|
|» fans_list|[object]|false|none||none|
|»» created_at|string|true|none||none|
|»» deleted_at|string|true|none||none|
|»» password|string|true|none||none|
|»» photo_url|string|true|none||none|
|»» updated_at|string|true|none||none|
|»» user_id|string|true|none||none|
|»» username|string|true|none||none|

<a id="opIdcommunicationGetSubscriberListHandler"></a>

## GET GetSubscriberListHandler

GET /following/list

> Body 请求参数

```yaml
page_number: "1"
page_size: "1"

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|
|» page_number|body|string| 是 |none|
|» page_size|body|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "base_response": {
    "status_code": 0,
    "status_msg": "string"
  },
  "subscriber_count": 0,
  "subscriber_list": [
    {
      "created_at": "string",
      "deleted_at": "string",
      "password": "string",
      "photo_url": "string",
      "updated_at": "string",
      "user_id": "string",
      "username": "string"
    }
  ]
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» base_response|object|false|none||none|
|»» status_code|integer|true|none||none|
|»» status_msg|string|true|none||none|
|» subscriber_count|integer|false|none||none|
|» subscriber_list|[object]|false|none||none|
|»» created_at|string|true|none||none|
|»» deleted_at|string|true|none||none|
|»» password|string|true|none||none|
|»» photo_url|string|true|none||none|
|»» updated_at|string|true|none||none|
|»» user_id|string|true|none||none|
|»» username|string|true|none||none|

<a id="opIdcommunicationGetFriendListHandler"></a>

## GET GetFriendListHandler

GET /friend/list

> Body 请求参数

```yaml
page_number: "0"
page_size: "0"

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|
|» page_number|body|string| 否 |none|
|» page_size|body|string| 否 |none|

> 返回示例

> 200 Response

```json
{
  "base_response": {
    "status_code": 0,
    "status_msg": "string"
  },
  "friend_count": 0,
  "friend_list": [
    {
      "created_at": "string",
      "deleted_at": "string",
      "password": "string",
      "photo_url": "string",
      "updated_at": "string",
      "user_id": "string",
      "username": "string"
    }
  ]
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» base_response|object|false|none||none|
|»» status_code|integer|true|none||none|
|»» status_msg|string|true|none||none|
|» friend_count|integer|false|none||none|
|» friend_list|[object]|false|none||none|
|»» created_at|string|true|none||none|
|»» deleted_at|string|true|none||none|
|»» password|string|true|none||none|
|»» photo_url|string|true|none||none|
|»» updated_at|string|true|none||none|
|»» user_id|string|true|none||none|
|»» username|string|true|none||none|

<a id="opIdcommunicationSubscribeHandler"></a>

## POST SubscribeHandler

POST /relation/action

> Body 请求参数

```yaml
to_user_id: u2046601598389456896
action_type: 0

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|
|» to_user_id|body|string| 是 |none|
|» action_type|body|integer| 是 |none|

> 返回示例

> 200 Response

```json
{
  "base_response": {
    "status_code": 0,
    "status_msg": "string"
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» base_response|object|false|none||none|
|»» status_code|integer|true|none||none|
|»» status_msg|string|true|none||none|

# user

<a id="opIduserPostUserPhotoHandler"></a>

## PUT PostUserPhotoHandler

PUT /user/avatar/upload

> Body 请求参数

```yaml
photo_url: file://D:\vsc-code\go_zero-tiktok\testdata\user1\image_570131856141295.png

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|
|» photo_url|body|string(binary)| 是 |none|

> 返回示例

> 200 Response

```json
{
  "status_code": 0,
  "status_msg": "string"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» status_code|integer|false|none||none|
|» status_msg|string|false|none||none|

<a id="opIduserGetUserInfoHandler"></a>

## GET GetUserInfoHandler

GET /user/info

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|

> 返回示例

> 200 Response

```json
{
  "base": {
    "status_code": 0,
    "status_msg": "string"
  },
  "user": {
    "created_at": "string",
    "deleted_at": "string",
    "password": "string",
    "photo_url": "string",
    "updated_at": "string",
    "user_id": "string",
    "username": "string"
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» base|object|false|none||none|
|»» status_code|integer|true|none||none|
|»» status_msg|string|true|none||none|
|» user|object|false|none||none|
|»» created_at|string|true|none||none|
|»» deleted_at|string|true|none||none|
|»» password|string|true|none||none|
|»» photo_url|string|true|none||none|
|»» updated_at|string|true|none||none|
|»» user_id|string|true|none||none|
|»» username|string|true|none||none|

<a id="opIduserLoginHandler"></a>

## POST LoginHandler

POST /user/login

> Body 请求参数

```yaml
username: ""
password: ""

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|
|» username|body|string| 是 |none|
|» password|body|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "access_token": "string",
  "base": {
    "status_code": 0,
    "status_msg": "string"
  },
  "refresh_token": "string",
  "user_id": "string"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» access_token|string|false|none||none|
|» base|object|false|none||none|
|»» status_code|integer|true|none||none|
|»» status_msg|string|true|none||none|
|» refresh_token|string|false|none||none|
|» user_id|string|false|none||none|

<a id="opIduserRegisterHandler"></a>

## POST RegisterHandler

POST /user/register

> Body 请求参数

```yaml
username: test3
password: "123456"

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|
|» username|body|string| 是 |none|
|» password|body|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "base": {
    "status_code": 0,
    "status_msg": "string"
  },
  "user_id": "string"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» base|object|false|none||none|
|»» status_code|integer|true|none||none|
|»» status_msg|string|true|none||none|
|» user_id|string|false|none||none|

<a id="opIduserRefreshTokenHandler"></a>

## POST RefreshTokenHandler

POST /user/token/refresh

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|

> 返回示例

> 200 Response

```json
{
  "access_token": "string",
  "base": {
    "status_code": 0,
    "status_msg": "string"
  },
  "refresh_token": "string"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» access_token|string|false|none||none|
|» base|object|false|none||none|
|»» status_code|integer|true|none||none|
|»» status_msg|string|true|none||none|
|» refresh_token|string|false|none||none|

# video

<a id="opIdvideoGetVideoListHandler"></a>

## GET GetVideoListHandler

GET /video/list

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|user_id|query|string| 否 |none|
|page_num|query|string| 否 |none|
|page_size|query|string| 否 |none|
|Authorization|header|string| 否 |none|

> 返回示例

> 200 Response

```json
{
  "base": {
    "status_code": 0,
    "status_msg": "string"
  },
  "videos": [
    {
      "author_id": "string",
      "cover_url": "string",
      "created_at": "string",
      "deleted_at": "string",
      "description": "string",
      "title": "string",
      "updated_at": "string",
      "video_id": "string",
      "video_url": "string"
    }
  ]
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» base|object|false|none||none|
|»» status_code|integer|true|none||none|
|»» status_msg|string|true|none||none|
|» videos|[object]|false|none||none|
|»» author_id|string|true|none||none|
|»» cover_url|string|true|none||none|
|»» created_at|string|true|none||none|
|»» deleted_at|string|true|none||none|
|»» description|string|true|none||none|
|»» title|string|true|none||none|
|»» updated_at|string|true|none||none|
|»» video_id|string|true|none||none|
|»» video_url|string|true|none||none|

<a id="opIdvideoVideoPopularHandler"></a>

## GET VideoPopularHandler

GET /video/popular

> Body 请求参数

```yaml
page_size: "10"
page_num: "1"

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|
|» page_size|body|string| 否 |none|
|» page_num|body|string| 否 |none|

> 返回示例

> 200 Response

```json
{
  "base": {
    "status_code": 0,
    "status_msg": "string"
  },
  "videos": [
    {
      "author_id": "string",
      "cover_url": "string",
      "created_at": "string",
      "deleted_at": "string",
      "description": "string",
      "title": "string",
      "updated_at": "string",
      "video_id": "string",
      "video_url": "string"
    }
  ]
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» base|object|false|none||none|
|»» status_code|integer|true|none||none|
|»» status_msg|string|true|none||none|
|» videos|[object]|false|none||none|
|»» author_id|string|true|none||none|
|»» cover_url|string|true|none||none|
|»» created_at|string|true|none||none|
|»» deleted_at|string|true|none||none|
|»» description|string|true|none||none|
|»» title|string|true|none||none|
|»» updated_at|string|true|none||none|
|»» video_id|string|true|none||none|
|»» video_url|string|true|none||none|

<a id="opIdvideoPublishVideoHandler"></a>

## POST PublishVideoHandler

POST /video/publish

> Body 请求参数

```yaml
title: 你好
description: hello
video_file: file://D:\vsc-code\go_zero-tiktok\testdata\user1\video_1776329938386_ny65e1.mp4

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|
|» title|body|string| 是 |none|
|» description|body|string| 是 |none|
|» video_file|body|string(binary)| 否 |none|

> 返回示例

> 200 Response

```json
{
  "base": {
    "status_code": 0,
    "status_msg": "string"
  },
  "video_id": "string"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» base|object|false|none||none|
|»» status_code|integer|true|none||none|
|»» status_msg|string|true|none||none|
|» video_id|string|false|none||none|

<a id="opIdvideoVideoSearchHandler"></a>

## GET VideoSearchHandler

GET /video/search

> Body 请求参数

```yaml
keyword: 你好
page_size: "0"
page_num: "0"

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|
|» keyword|body|string| 否 |none|
|» page_size|body|string| 否 |none|
|» page_num|body|string| 否 |none|

> 返回示例

> 200 Response

```json
{
  "base": {
    "status_code": 0,
    "status_msg": "string"
  },
  "videos": [
    {
      "author_id": "string",
      "cover_url": "string",
      "created_at": "string",
      "deleted_at": "string",
      "description": "string",
      "title": "string",
      "updated_at": "string",
      "video_id": "string",
      "video_url": "string"
    }
  ]
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» base|object|false|none||none|
|»» status_code|integer|true|none||none|
|»» status_msg|string|true|none||none|
|» videos|[object]|false|none||none|
|»» author_id|string|true|none||none|
|»» cover_url|string|true|none||none|
|»» created_at|string|true|none||none|
|»» deleted_at|string|true|none||none|
|»» description|string|true|none||none|
|»» title|string|true|none||none|
|»» updated_at|string|true|none||none|
|»» video_id|string|true|none||none|
|»» video_url|string|true|none||none|

# 数据模型

<h2 id="tocS_CommentBaseinfos">CommentBaseinfos</h2>

<a id="schemacommentbaseinfos"></a>
<a id="schema_CommentBaseinfos"></a>
<a id="tocScommentbaseinfos"></a>
<a id="tocscommentbaseinfos"></a>

```json
{
  "comment_id": "string",
  "user_id": "string",
  "video_id": "string",
  "content": "string",
  "created_at": "string",
  "updated_at": "string",
  "deleted_at": "string",
  "like_count": "0"
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|comment_id|string|true|none||none|
|user_id|string|true|none||none|
|video_id|string|true|none||none|
|content|string|true|none||none|
|created_at|string¦null|false|none||none|
|updated_at|string¦null|false|none||none|
|deleted_at|string¦null|false|none||none|
|like_count|integer¦null|false|none||none|

<h2 id="tocS_UserBaseinfos">UserBaseinfos</h2>

<a id="schemauserbaseinfos"></a>
<a id="schema_UserBaseinfos"></a>
<a id="tocSuserbaseinfos"></a>
<a id="tocsuserbaseinfos"></a>

```json
{
  "user_id": "string",
  "username": "string",
  "password": "string",
  "photo_url": "string",
  "created_at": "string",
  "updated_at": "string",
  "deleted_at": "string"
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|user_id|string|true|none||none|
|username|string¦null|false|none||none|
|password|string|true|none||none|
|photo_url|string|true|none||none|
|created_at|string¦null|false|none||none|
|updated_at|string¦null|false|none||none|
|deleted_at|string¦null|false|none||none|

<h2 id="tocS_UserFollows">UserFollows</h2>

<a id="schemauserfollows"></a>
<a id="schema_UserFollows"></a>
<a id="tocSuserfollows"></a>
<a id="tocsuserfollows"></a>

```json
{
  "follower_id": "string",
  "user_id": "string",
  "status": "0"
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|follower_id|string|true|none||none|
|user_id|string|true|none||none|
|status|integer¦null|false|none||none|

<h2 id="tocS_VideoBaseinfos">VideoBaseinfos</h2>

<a id="schemavideobaseinfos"></a>
<a id="schema_VideoBaseinfos"></a>
<a id="tocSvideobaseinfos"></a>
<a id="tocsvideobaseinfos"></a>

```json
{
  "video_id": "string",
  "author_id": "string",
  "video_url": "string",
  "cover_url": "string",
  "title": "string",
  "description": "string",
  "created_at": "string",
  "updated_at": "string",
  "deleted_at": "string"
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|video_id|string|true|none||none|
|author_id|string|true|none||none|
|video_url|string|true|none||none|
|cover_url|string¦null|false|none||none|
|title|string|true|none||none|
|description|string¦null|false|none||none|
|created_at|string¦null|false|none||none|
|updated_at|string¦null|false|none||none|
|deleted_at|string¦null|false|none||none|

<h2 id="tocS_VideoLikers">VideoLikers</h2>

<a id="schemavideolikers"></a>
<a id="schema_VideoLikers"></a>
<a id="tocSvideolikers"></a>
<a id="tocsvideolikers"></a>

```json
{
  "user_id": "string",
  "video_id": "string"
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|user_id|string|true|none||none|
|video_id|string|true|none||none|

<h2 id="tocS_VideoPopulars">VideoPopulars</h2>

<a id="schemavideopopulars"></a>
<a id="schema_VideoPopulars"></a>
<a id="tocSvideopopulars"></a>
<a id="tocsvideopopulars"></a>

```json
{
  "video_id": "string",
  "visit_count": "0",
  "like_count": "0",
  "comment_count": "0"
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|video_id|string|true|none||none|
|visit_count|integer¦null|false|none||none|
|like_count|integer¦null|false|none||none|
|comment_count|integer¦null|false|none||none|

