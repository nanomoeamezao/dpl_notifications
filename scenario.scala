package wstest

import scala.concurrent.duration._

import io.gatling.core.Predef._
import io.gatling.http.Predef._
import io.gatling.http.protocol.HttpProtocolBuilder

class Websocket extends Simulation {
  val httpProtocol: HttpProtocolBuilder = http.baseUrl("192.168.0.100").wsBaseUrl("ws://192.168.0.100/ws")

//root is HTTP protocol and webSocket url is being appended to it

  val scene = scenario("testWebSocket")
    .exec(http("firstRequest")
    .get("/"))
    .exec(addCookie(Cookie("UID", "111")))
    .exec(ws("openSocket").connect("/"))
    .pause(20)
    .exec(ws("closeConnection").close)
//terminating the current websocket connection

  setUp(scene.inject(
    atOnceUsers(20)
    rampUsers(100).during(20.seconds)
    ).protocols(httpProtocol))
}