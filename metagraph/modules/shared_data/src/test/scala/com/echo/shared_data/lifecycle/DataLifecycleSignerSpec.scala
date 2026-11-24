package com.echo.shared_data.lifecycle

import org.scalatest.flatspec.AnyFlatSpec
import org.scalatest.matchers.should.Matchers

class DataLifecycleSignerSpec extends AnyFlatSpec with Matchers {

  private val gateway  = "a" * 128
  private val stranger = "b" * 128

  "parseAuthorizedSigners" should "treat unset, empty and blank lists as no restriction" in {
    DataLifecycle.parseAuthorizedSigners(None) shouldBe None
    DataLifecycle.parseAuthorizedSigners(Some("")) shouldBe None
    DataLifecycle.parseAuthorizedSigners(Some(" , ")) shouldBe None
  }

  it should "trim and lower-case the listed keys" in {
    DataLifecycle.parseAuthorizedSigners(Some(s" ${gateway.toUpperCase} ,$stranger")) shouldBe
      Some(Set(gateway, stranger))
  }

  "checkSigners" should "accept any signer when no list is configured" in {
    DataLifecycle.checkSigners(List(stranger), None) shouldBe Right(())
  }

  it should "accept an authorised signer regardless of hex case" in {
    DataLifecycle.checkSigners(List(gateway.toUpperCase), Some(Set(gateway))) shouldBe Right(())
  }

  it should "reject a signer that is not on the list" in {
    DataLifecycle.checkSigners(List(stranger), Some(Set(gateway))).isLeft shouldBe true
  }

  it should "reject when any one of several proofs is not authorised" in {
    DataLifecycle.checkSigners(List(gateway, stranger), Some(Set(gateway))).isLeft shouldBe true
  }

  it should "reject an update with no proofs when a list is configured" in {
    DataLifecycle.checkSigners(Nil, Some(Set(gateway))).isLeft shouldBe true
  }
}
