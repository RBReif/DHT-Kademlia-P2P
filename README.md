We implement a distributed hashtable following the Kademlia approach (https://pdos.csail.mit.edu/~petar/papers/maymounkov-kademlia-lncs.pdf).
This project was developed by Manfred Stoiber and Roland Reif as part of the course "P2P Systems and Security" at TU Munich.

We have separated our implementation into three main modules, namely the API communication, the peerto-peer communication and the k-Buckets. The API communication listens for newly to be established
connections and listens on these connections for new orders for our distributed hash-table. These orders can
be of type dhtPUT and dhtGET. As an answer, a dhtSUCCESS and dhtFAILURE message can be received.
The peer-to-peer module is responsible for parsing the messages we use for peer-to-peer communication and
for the logic of how to respond to specific P2P messages.
k-Buckets are an implementation specific datastructure for storing known peers. The k-Buckets module
manages these k-buckets e.g. populates the k-buckets and updates them.