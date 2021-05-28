package wstest

import scala.concurrent.duration._

import io.gatling.core.Predef._
import io.gatling.http.Predef._
import io.gatling.http.protocol.HttpProtocolBuilder

class Websocket extends Simulation {
  val httpProtocol: HttpProtocolBuilder = http.baseUrl("http://192.168.0.100:8080").wsBaseUrl("ws://localhost:8080")

//root is HTTP protocol and webSocket url is being appended to it

  val scene = scenario("testWebSocket")
    .exec(http("firstRequest")
    .get("/"))
    .exec(addCookie(Cookie("UID", "111").withDomain("localhost")))
    .exec(addCookie(Cookie("lastID", "0").withDomain("localhost")))
    .pause(1)
    .exec(ws("openSocket").connect("/ws"))
    .pause(30.seconds)
    .exec(ws("closeConnection").close)
//terminating the current websocket connection

  setUp(scene.inject(
    atOnceUsers(20),
    rampUsers(200).during(20.seconds)
    ).protocols(httpProtocol))
}