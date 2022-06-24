package main

const (
	loginPage = `<form action="" method="post">
	<label for="username">Username: </label><input type="text" name="username" value="{{.Username}}" /><span>{{.Error}}</span><br />
	<label for="password">Password: </label><input type="password" name="password" /><br />
	<input type="submit" value="Login" />
</form>`
	adminPage = `<!doctype html>
<html lang="en">
	<head>
		<title>Admin</title>
		<script type="module" src="admin.js"></script>
	</head>
	<body>
		<h1>Loading</h1>
	</body>
</html>
`
)
