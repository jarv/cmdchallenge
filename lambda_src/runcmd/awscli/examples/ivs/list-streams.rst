**To get a list of live streams and their state**

The following ``list-streams`` example lists all live streams for your AWS account. ::

    aws ivs list-streams

Output::

    {
       "streams": [
            {
                "channelArn": "arn:aws:ivs:us-west-2:123456789012:channel/abcdABCDefgh",
                "state": "LIVE",
                "health": "HEALTHY",
                "viewerCount": 1
            }
        ]
    }

For more information, see `Create a Channel <https://docs.aws.amazon.com/ivs/latest/userguide/GSIVS-create-channel.html>`__ in the *Amazon Interactive Video Service User Guide*.