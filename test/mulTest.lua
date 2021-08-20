request = function()
    user = math.random(1, 10000000)
    wrk.method = "POST"
    wrk.body = "username=user" .. user .. "&password=Password_" .. user
    wrk.headers["Content-Type"] = "application/x-www-form-urlencoded"
    return wrk.format()
end