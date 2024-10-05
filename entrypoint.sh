#!/bin/sh

# 1 value を取得 : Not Found
curl -X GET localhost:8080

# 2 value を取得 : Not Found
curl -X GET localhost:9090

# 3 value を登録 : value
curl -X POST localhost:8080

# 4 value を取得 ( from redis ) : value
curl -X GET localhost:8080

# 5 value を取得 ( from in memory ) : value
curl -X GET localhost:8080

# 6 value を取得 ( from redis ) : value
curl -X GET localhost:9090

# 7 value を取得 ( from in memory ) : value
curl -X GET localhost:9090

# 8 value を9090に更新 : 9090
curl -X PUT localhost:9090

# 9 value を取得 ( from in memory ) : 9090
curl -X GET localhost:9090

# 10 value を取得 ( from in memory ) : 9090
curl -X GET localhost:8080

# 11 value を削除
curl -X DELETE localhost:8080

# 12 value を取得 : not found
curl -X GET localhost:8080

# 13 value を取得 : not found
curl -X GET localhost:9090
