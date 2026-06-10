package com.echo.shared_data.state

import com.echo.shared_data.types.MerkleRootStatus

import java.util.concurrent.ConcurrentHashMap
import scala.jdk.CollectionConverters._

/** In-memory index for WO-230 Step 5 finality polling on Data L1. */
object DataMerkleRootIndex {

  private val roots = new ConcurrentHashMap[String, MerkleRootStatus]()

  def record(root: String, leafCount: Int, finalized: Boolean): Unit = {
    roots.put(
      root,
      MerkleRootStatus(root = root, leafCount = leafCount, finalized = finalized)
    )
    ()
  }

  def lookup(root: String): Option[MerkleRootStatus] =
    Option(roots.get(root))

  def clear(): Unit =
    roots.clear()

  def snapshot: Map[String, MerkleRootStatus] =
    roots.asScala.toMap
}
