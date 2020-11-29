**To get information about a stream**

The following ``get-stream-key`` example gets information about the specified stream key. ::

    aws ivs get-stream \
        --channel-arn arn:aws:ivs:us-west-2:123456789012:stream-key/skSKABCDefgh

Output::

    {
        "streamKey": {
            "arn": "arn:aws:ivs:us-west-2:123456789012:stream-key/skSKABCDefgh",
            "value": "sk_us-west-2_abcdABCDefgh_567890abcdef",
            "channelArn": "arn:aws:ivs:us-west-2:123456789012:channel/abcdABCDefgh",
            "tags": {}
        }
    }

For more information, see `Create a Channel <https://docs.aws.amazon.com/ivs/latest/userguide/GSIVS-create-channel.html>`__ in the *Amazon Interactive Video Service User Guide*.