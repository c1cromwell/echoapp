package com.echo.shared_data.errors

import cats.syntax.all._
import io.constellationnetwork.currency.dataApplication.DataApplicationValidationError
import io.constellationnetwork.currency.dataApplication.dataApplication.DataApplicationValidationErrorOr

object DataErrors {
  type Result = DataApplicationValidationErrorOr[Unit]

  val valid: Result = ().validNec[DataApplicationValidationError]

  final case class Rejected(reason: String) extends DataApplicationValidationError {
    val message: String = reason
  }

  def invalid(reason: String): Result =
    Rejected(reason).invalidNec[Unit]
}
