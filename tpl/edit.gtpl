<!DOCTYPE html>
<html>
<body>
  <div>
    <h1>Edit Profile</h1>
    </br>
    <form enctype="multipart/form-data" action="http://alejandroseaah.com:9090/upload" method="post">
        <label for="nname"><b>nickname :</b></label></br>
        <input type="text" name="nickname" />
        </br>
        </br>
        <label for="avatar"><b>upload avatar :</b></label></br>
        <input type="file" name="uploadfile" />
        <input type="hidden" name="token" value="{{.}}"/></br>
        <button type="button">Cancel</button>
        <input type="submit" value="Save" />        
    </form>
  </div>
</body>
</html>
