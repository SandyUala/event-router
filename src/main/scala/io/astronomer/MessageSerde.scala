package io.astronomer

import java.util

import org.apache.kafka.common.serialization.{Deserializer, Serde, Serializer}

class MessageDeserializer extends Deserializer[Message] {
  override def configure(configs: util.Map[String, _], isKey: Boolean): Unit = ()

  override def close(): Unit = ()

  override def deserialize(topic: String, data: Array[Byte]): Message = {
    val m = Json.ByteArray.decode[Message](data)
    m.raw = data
    m
  }
}

class MessageSerializer extends Serializer[Message] {
  override def configure(configs: util.Map[String, _], isKey: Boolean): Unit = ()

  override def serialize(topic: String, data: Message): Array[Byte] = data.raw

  override def close(): Unit = ()
}

class MessageSerde extends Serde[Message] {
  override def deserializer(): Deserializer[Message] = new MessageDeserializer()

  override def configure(configs: util.Map[String, _], isKey: Boolean): Unit = ()

  override def close(): Unit = ()

  override def serializer(): Serializer[Message] = new MessageSerializer()
}
