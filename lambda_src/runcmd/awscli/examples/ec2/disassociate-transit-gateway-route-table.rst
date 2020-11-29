**To disassociate a transit gateway route table from a  resource attachment**

The following ``disassociate-transit-gateway-route-table`` example disasssociates the transit gateway route table from the specified attachment. ::

    aws ec2 disassociate-transit-gateway-route-table \
        --transit-gateway-route-table-id tgw-rtb-002573ed1eEXAMPLE \
        --transit-gateway-attachment-id tgw-attach-08e0bc912cEXAMPLE

Output::

    {
        "Association": {
            "TransitGatewayRouteTableId": "tgw-rtb-002573ed1eEXAMPLE",
            "TransitGatewayAttachmentId": "tgw-attach-08e0bc912cEXAMPLE",
            "ResourceId": "11460968-4ac1-4fd3-bdb2-00599EXAMPLE",
            "ResourceType": "direct-connect-gateway",
            "State": "disassociating"
        }
    }

For more information, see `Delete an Association for a Transit Gateway Route Table <https://docs.aws.amazon.com/vpc/latest/tgw/tgw-route-tables.html#disassociate-tgw-route-table>`__ in the *AWS Transit Gateways Guide*.
