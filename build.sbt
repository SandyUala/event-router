import Dependencies._

lazy val root = (project in file(".")).
  settings(
    inThisBuild(List(
      organization := "com.example",
      scalaVersion := "2.12.3",
      version      := "0.1.0-SNAPSHOT"
    )),
    name := "Hello",
    libraryDependencies ++= Seq(
      scalaTest % Test,
      "org.apache.kafka" % "kafka-clients" % "0.11.0.0",
      "org.apache.kafka" % "kafka-streams" % "0.11.0.0",
      "com.fasterxml.jackson.module" % "jackson-module-scala_2.12" % "2.9.1"
    )
  )
