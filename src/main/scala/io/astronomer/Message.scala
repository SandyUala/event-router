package io.astronomer

import com.fasterxml.jackson.annotation.JsonIgnoreProperties

import scala.collection.mutable.ArrayBuffer

@JsonIgnoreProperties(ignoreUnknown = true)
case class Message(appId: Option[String], integrations: Option[Map[String, Boolean]]) {
  var raw: Array[Byte] = _
  var integrationsToSend: ArrayBuffer[String] = new ArrayBuffer[String]

  override def toString: String = {
    s"appId => ${appId.get}"
  }
}