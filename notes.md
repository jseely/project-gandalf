# For Manisha

Just a few things I wanted to point out about the new APIs for group membership and authentication.

They now both require the initial *Request that was sent to project gandalf along with the ResponseWriter.
They both return a collection of headers and an error.
If they return (nil, nil) assume they have written a redirect request to the ResponseWriter and just return from your handler without any further action.
If they return (http.Header, nil) assume that everything worked correctly and append all the headers listed to the initial request and the request that will be forwarded.
If they return (nil, error) assume there was an error performing the action and write an InternalServerError to the ResponseWriter.
