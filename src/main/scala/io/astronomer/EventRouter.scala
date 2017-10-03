package io.astronomer

import java.util.Properties

import org.apache.kafka.common.serialization.Serdes
import org.apache.kafka.streams.kstream.{KStreamBuilder, Predicate}
import org.apache.kafka.streams.{KafkaStreams, StreamsConfig}

import scala.collection.mutable
import scala.collection.mutable.ArrayBuffer

object EventRouter extends App {
  val builder = new KStreamBuilder

  val streamingConfig = {
    val settings = new Properties
    settings.put(StreamsConfig.APPLICATION_ID_CONFIG, "scala-test")
    settings.put(StreamsConfig.BOOTSTRAP_SERVERS_CONFIG, "192.168.99.100:9092")
    settings
  }

  val messages = builder.stream(Serdes.String, new MessageSerde, "main")

  messages.print()

  // get globally enabled integrations from API?
  val integrations = Array[String](
    "mixpanel",
    "google-analytics",
  )

  println(integrations.mkString(", "))

  def predicateFactory (integration: String): Predicate[String, Message] = {
    (k: String, v: Message) => v.integrationsToSend.contains(integration)
  }

  val predicates = integrations.map(predicateFactory)

  // cache of appIds with enabled integrations
  var cache = mutable.Map[String, ArrayBuffer[String]]()
  cache += ("2xSYb9bncaMAppvvwxRpe" -> ArrayBuffer[String](
      "mixpanel",
      "google-analytics",
    )
  )

  messages
    .filter((_, value: Message) => value.appId.isDefined)
    .mapValues((m: Message) => {
      m.integrationsToSend = cache(m.appId.get)
      // do more logic here to handle integrations that are enabled/disabled
      // on this message, as well as not forwarding for client side clients
      m
    })

  predicates.zipWithIndex.foreach {
    case(p: Predicate[String, Message], index: Int) => {
      messages.filter(p).to(Serdes.String, new MessageSerde, integrations(index))
    }
  }

  val stream: KafkaStreams = new KafkaStreams(builder, streamingConfig)
  stream.start()
}
