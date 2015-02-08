#[Static Contact](http://www.staticcontact.com)

A form forwarder. Static Contact can replace the backend of a contact form so that you can put one on a static site and receive the email in your inbox.

##How-to

Simply change the action and method of a form to be `http://www.staticcontact.com/{your_email}` and `POST`, respectively.

Here is a sample form:

``` HTML
<form id="contact-form" action="http://www.staticcontact.com/contact/{your_email}" method="POST">
    <input type="text" name="name" placeholder="name" required/><br/>
    <input type="text" name="email" placeholder="email" required/><br/>
    <input type="text" name="subject" placeholder="subject"/><br/>
    <input type="text" name="message" placeholder="message" required/><br/>
    <input id="contact-submit" type="submit" value="Say hello!" />
</form>
```

Typically, a form submission redirects a user. The below javascript submits the form but does not redirect the user.

``` js
$(function() {
  $("#contact-form").on("submit", function(e) {
    e.preventDefault();
    $.ajax({
        url: $(this).attr("action"),
        type: 'POST',
        data: $(this).serialize(),
        success: function(data) {
            $("#contact-submit").val('Form successfully submitted!')
        }
    });
  });
});
```

## Roll your own

It is quite easy to deploy your own version of this. What is required is a Heroku and Mandrill account. This project has a connection to a MongoLab database - this was going to be used to store sends per month for a given user if the volume were to approach the 12,000 Mandrill maximum.
