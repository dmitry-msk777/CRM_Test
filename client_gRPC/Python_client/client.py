import grpc
import CRMswap_pb2
import CRMswap_pb2_grpc
# открываем канал и создаем клиент
channel = grpc.insecure_channel('localhost:5300')
stub = CRMswap_pb2_grpc.CRMswapStub(channel)
# запрос
RequestGET_query = CRMswap_pb2.RequestGET(Customer_id='666')
response = stub.GET_List(RequestGET_query)
print(response)