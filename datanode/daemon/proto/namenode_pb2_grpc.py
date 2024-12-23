# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc
import warnings

from . import namenode_pb2 as namenode__pb2

GRPC_GENERATED_VERSION = '1.68.1'
GRPC_VERSION = grpc.__version__
_version_not_supported = False

try:
    from grpc._utilities import first_version_is_lower
    _version_not_supported = first_version_is_lower(GRPC_VERSION, GRPC_GENERATED_VERSION)
except ImportError:
    _version_not_supported = True

if _version_not_supported:
    raise RuntimeError(
        f'The grpc package installed is at version {GRPC_VERSION},'
        + f' but the generated code in namenode_pb2_grpc.py depends on'
        + f' grpcio>={GRPC_GENERATED_VERSION}.'
        + f' Please upgrade your grpc module to grpcio>={GRPC_GENERATED_VERSION}'
        + f' or downgrade your generated code using grpcio-tools<={GRPC_VERSION}.'
    )


class NameServiceStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.CreateObject = channel.unary_unary(
                '/proto.NameService/CreateObject',
                request_serializer=namenode__pb2.CreateObjectRequest.SerializeToString,
                response_deserializer=namenode__pb2.CreateObjectResponse.FromString,
                _registered_method=True)
        self.DeleteObject = channel.unary_unary(
                '/proto.NameService/DeleteObject',
                request_serializer=namenode__pb2.DeleteObjectRequest.SerializeToString,
                response_deserializer=namenode__pb2.DeleteObjectResponse.FromString,
                _registered_method=True)
        self.UpdateObject = channel.unary_unary(
                '/proto.NameService/UpdateObject',
                request_serializer=namenode__pb2.UpdateObjectReq.SerializeToString,
                response_deserializer=namenode__pb2.UpdateObjectRes.FromString,
                _registered_method=True)
        self.LeaseObject = channel.unary_unary(
                '/proto.NameService/LeaseObject',
                request_serializer=namenode__pb2.LeaseObjectReq.SerializeToString,
                response_deserializer=namenode__pb2.LeaseObjectRes.FromString,
                _registered_method=True)


class NameServiceServicer(object):
    """Missing associated documentation comment in .proto file."""

    def CreateObject(self, request, context):
        """this service defines procedures to be used for the object store operations
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def DeleteObject(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def UpdateObject(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def LeaseObject(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_NameServiceServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'CreateObject': grpc.unary_unary_rpc_method_handler(
                    servicer.CreateObject,
                    request_deserializer=namenode__pb2.CreateObjectRequest.FromString,
                    response_serializer=namenode__pb2.CreateObjectResponse.SerializeToString,
            ),
            'DeleteObject': grpc.unary_unary_rpc_method_handler(
                    servicer.DeleteObject,
                    request_deserializer=namenode__pb2.DeleteObjectRequest.FromString,
                    response_serializer=namenode__pb2.DeleteObjectResponse.SerializeToString,
            ),
            'UpdateObject': grpc.unary_unary_rpc_method_handler(
                    servicer.UpdateObject,
                    request_deserializer=namenode__pb2.UpdateObjectReq.FromString,
                    response_serializer=namenode__pb2.UpdateObjectRes.SerializeToString,
            ),
            'LeaseObject': grpc.unary_unary_rpc_method_handler(
                    servicer.LeaseObject,
                    request_deserializer=namenode__pb2.LeaseObjectReq.FromString,
                    response_serializer=namenode__pb2.LeaseObjectRes.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'proto.NameService', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))
    server.add_registered_method_handlers('proto.NameService', rpc_method_handlers)


 # This class is part of an EXPERIMENTAL API.
class NameService(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def CreateObject(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/proto.NameService/CreateObject',
            namenode__pb2.CreateObjectRequest.SerializeToString,
            namenode__pb2.CreateObjectResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def DeleteObject(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/proto.NameService/DeleteObject',
            namenode__pb2.DeleteObjectRequest.SerializeToString,
            namenode__pb2.DeleteObjectResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def UpdateObject(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/proto.NameService/UpdateObject',
            namenode__pb2.UpdateObjectReq.SerializeToString,
            namenode__pb2.UpdateObjectRes.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def LeaseObject(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/proto.NameService/LeaseObject',
            namenode__pb2.LeaseObjectReq.SerializeToString,
            namenode__pb2.LeaseObjectRes.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)


class DataServiceStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.RegisterNode = channel.stream_stream(
                '/proto.DataService/RegisterNode',
                request_serializer=namenode__pb2.NodeHeartBeat.SerializeToString,
                response_deserializer=namenode__pb2.CommandNodeRes.FromString,
                _registered_method=True)


class DataServiceServicer(object):
    """Missing associated documentation comment in .proto file."""

    def RegisterNode(self, request_iterator, context):
        """this service defines procedures to be used by the namenode to register and manage datanodes
        the initial register request sends a heartbeat with the request
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_DataServiceServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'RegisterNode': grpc.stream_stream_rpc_method_handler(
                    servicer.RegisterNode,
                    request_deserializer=namenode__pb2.NodeHeartBeat.FromString,
                    response_serializer=namenode__pb2.CommandNodeRes.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'proto.DataService', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))
    server.add_registered_method_handlers('proto.DataService', rpc_method_handlers)


 # This class is part of an EXPERIMENTAL API.
class DataService(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def RegisterNode(request_iterator,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.stream_stream(
            request_iterator,
            target,
            '/proto.DataService/RegisterNode',
            namenode__pb2.NodeHeartBeat.SerializeToString,
            namenode__pb2.CommandNodeRes.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)
