package com.echo.shared_data.routes

import cats.effect.IO
import com.echo.shared_data.state.DataMerkleRootIndex
import io.circe.syntax.EncoderOps
import org.http4s.circe.CirceEntityCodec.circeEntityEncoder
import org.http4s.dsl.io._
import org.http4s.{HttpRoutes, Method}

/** Read API for WO-230 Step 5 finality polling on Data L1. */
object DataRoutes {

  def merkleRootRoutes: HttpRoutes[IO] =
    HttpRoutes.of[IO] {
      case req @ GET -> Root / "data-application" / "merkle-roots" / root
          if req.method == Method.GET =>
        DataMerkleRootIndex.lookup(root) match {
          case Some(status) => Ok(status.asJson)
          case None         => NotFound()
        }
    }
}
