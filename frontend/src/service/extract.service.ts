import { Injectable } from '@angular/core';
import {Extract} from "../../wailsjs/go/services/ExtractService";

@Injectable({
  providedIn: 'root'
})
export class ExtractService {
  async extraction(file: string) {
    await Extract(file);
  }
}
