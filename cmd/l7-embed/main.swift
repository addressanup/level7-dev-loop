import Foundation
import NaturalLanguage

struct Request: Decodable {
    let schema: Int
    let texts: [String]
}

struct Response: Encodable {
    let schema: Int
    let revision: Int
    let dimension: Int
    let vectors: [[Double]]
}

do {
    let input = FileHandle.standardInput.readDataToEndOfFile()
    guard input.count > 1 && input.count <= 1_048_576 else {
        throw NSError(domain: "Level7Embedding", code: 1)
    }
    let request = try JSONDecoder().decode(Request.self, from: input)
    guard request.schema == 1 && !request.texts.isEmpty else {
        throw NSError(domain: "Level7Embedding", code: 2)
    }
    guard let embedding = NLEmbedding.sentenceEmbedding(for: .english) else {
        throw NSError(domain: "Level7Embedding", code: 3)
    }
    var vectors: [[Double]] = []
    for text in request.texts {
        guard !text.isEmpty && text.utf8.count <= 32_768,
              let vector = embedding.vector(for: text),
              vector.count == embedding.dimension else {
            throw NSError(domain: "Level7Embedding", code: 4)
        }
        vectors.append(vector)
    }
    let response = Response(schema: 1, revision: embedding.revision, dimension: embedding.dimension, vectors: vectors)
    FileHandle.standardOutput.write(try JSONEncoder().encode(response))
} catch {
    FileHandle.standardError.write(Data("embedding unavailable\n".utf8))
    exit(1)
}
