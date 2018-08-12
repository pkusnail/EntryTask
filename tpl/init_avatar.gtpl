<html>
    <head>
    <title>Welcome {{.nickname}}</title>
    </head>
    <body>
	Sign up successfully!
	continue to upload your profile pic 
        <form action="/login" method="post">
            Username:<input type="text" name="username">
            Password:<input type="password" name="password">
            <input type="submit" value="Login">
        </form>
    </body>
</html>

