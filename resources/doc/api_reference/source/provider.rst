Provider API
============

Available providers
------------------- 

.. http:get:: /v1/provider

    Retrieve a list all available providers.

    **Example request**:

    .. sourcecode:: http

        GET /v1/provider HTTP/1.1
        Host: example.com
        Authorization: dEFFEFeddedeGGEGMceokr353521234
        Accept: application/json

    **Example response**:

    .. sourcecode:: http

        HTTP/1.1 200 OK
        Accept: application/json
        Authorization: Basic cm9vdDpyb290
        Content-Type: application/json

        {
            "Status": "success",
            "Info": "providers found",
            "Response": [
                {
                    "ID": 1,
                    "Name": "fritzpay"
                }
            ],
            "Error": null
        }

    :statuscode 200: No error, provider datadata served.
    :statuscode 401: Unauthorized, either the username does not exist or the credentials

    :reqheader Authorization: A valid authorization token.


    .. note:: 
    
      This response is just an example, usually it is much longer!

Check provider
--------------

.. http:get:: /v1/provider/(providerID)

    Check if a specivic providerID is available in the system.

    **Example request**:

    .. sourcecode:: http

        GET /v1/provider/1 HTTP/1.1
        Host: example.com
        Authorization: dEFFEFeddedeGGEGMceokr353521234
        Accept: application/json

    :param providerID: string [0-9]

    **Example response**:

    .. sourcecode:: http

        HTTP/1.1 200 OK
        Accept: application/json
        Authorization: Basic cm9vdDpyb290
        Content-Type: application/json

        {
            "Status": "success",
            "Info": "provider fritzpay found",
            "Response": [
                {
                    "ID": 1,
                    "Name": "fritzpay"
                }
            ],
            "Error": null
        }

    :statuscode 200: No error, provider datadata served.
    :statuscode 400: The request was malformed; the given providerID could not be understood.
    :statuscode 401: Unauthorized, either the username does not exist or the credentials
    :statuscode 404: provider not available

    :reqheader Authorization: A valid authorization token.