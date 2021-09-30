# github-calendar-api
使用goquery爬取github用户贡献日历热力图数据,并使用gin接口化输出.
输出格式为:
````
"code": 2000,
"msg": "success",
"data": {
    "total": 1777,
    "contributions": [
        {
          "date": "2020-09-27",
        "count": 2,
        "level": 1
        },
        ...
        "date": "2021-09-30",
        "count": 0,
        "level": 0
    ],
 },
````
