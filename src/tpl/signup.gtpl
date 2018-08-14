<html>
    <head>
    <title>Welcome to sign up</title>
    </head>
    <body>
        <form action="/signup" method="post">
            Your real name:<input type="text" name="rname"> * use when login</br> 
            Your nick name:<input type="text" name="nname"></br>
            Password:<input type="password" name="pwd1"></br>
            Confirm Password:<input type="password" name="pwd2"></br>
            <input type="submit" value="Signin" onclick="return Validate()" >
        </form>

<script type="text/javascript">
window.onload = function () {
	document.getElementById("pwd1").onchange = validatePassword;
	document.getElementById("pwd2").onchange = validatePassword;
}
function validatePassword(){
var pass2=document.getElementById("pwd2").value;
var pass1=document.getElementById("pwd1").value;
if(pass1!=pass2)
	document.getElementById("pwd2").setCustomValidity("Passwords Don't Match");
else
	document.getElementById("pwd2").setCustomValidity('');	 
//empty string means no validation error
}
</script>

<script type="text/javascript">
    function Validate() {
        var pwd = document.getElementById("pwd1").value;
        var confirmPassword = document.getElementById("pwd2").value;
        if (pwd != confirmPassword) {
            alert("Passwords do not match.");
            return false;
        }
        return true;
    }
</script>

    </body>
</html>

